package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ==========================================
// ВАРИАНТ 1: Markdown (Идеально для Сегодня/Завтра/Одного дня)
// ==========================================

type RichMessageMarkdownRequest struct {
	ChatID      int64               `json:"chat_id"`
	RichMessage RichMessageMarkdown `json:"rich_message"`
}

type RichMessageMarkdown struct {
	Markdown string `json:"markdown"`
}

func BuildDayMarkdown(day DaySchedule, label string) string {
	title := fmt.Sprintf("### 📆 %s", day.Name)
	if label != "" {
		title = fmt.Sprintf("### 📆 %s (%s)", day.Name, label)
	}

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")
	sb.WriteString("| Пара | Занятие | Аудитория |\n| :--- | :--- | :--- |\n")
	for _, ev := range day.Events {
		sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", ev.Time, ev.Subject, ev.Location))
	}
	return sb.String()
}

func SendDayMarkdown(bot *tgbotapi.BotAPI, chatID int64, markdown string) error {
	payload := RichMessageMarkdownRequest{
		ChatID:      chatID,
		RichMessage: RichMessageMarkdown{Markdown: markdown},
	}
	return sendRichPayload(bot, payload)
}

// ==========================================
// ВАРИАНТ 2: Blocks (Идеально для Недели с выпадающими списками)
// ==========================================

type RichMessageBlocksRequest struct {
	ChatID      int64             `json:"chat_id"`
	RichMessage RichMessageBlocks `json:"rich_message"`
}

type RichMessageBlocks struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type       string   `json:"type"`
	Summary    string   `json:"summary,omitempty"` // Заголовок для details
	Blocks     []Block  `json:"blocks,omitempty"`  // Вложенные блоки для details
	Cells      [][]Cell `json:"cells,omitempty"`   // Ячейки для table
	IsBordered bool     `json:"is_bordered,omitempty"`
	IsStriped  bool     `json:"is_striped,omitempty"`
	Text       string   `json:"text,omitempty"` // Текст для paragraph (без Markdown!)
}

type Cell struct {
	Text string `json:"text"`
}

func BuildWeekBlocks() []Block {
	var blocks []Block

	// ✅ ИСПОЛЬЗУЕМ НАШУ ФУНКЦИЮ ВМЕСТО ISOWeek
	isEven := isEvenWeek()
	weekType := "нечётная"
	if isEven {
		weekType = "чётная"
	}

	// Вычисляем академический номер недели для красоты (от 1 сентября 2024)
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	daysPassed := int(time.Since(startDate).Hours() / 24)
	academicWeek := (daysPassed / 7) + 1

	// Общий заголовок недели
	blocks = append(blocks, Block{
		Type: "paragraph",
		Text: fmt.Sprintf("📆 Неделя (%s) • %d", weekType, academicWeek),
	})

	sched := GetSchedule()

	for _, dayKey := range DayOrder {
		day, exists := sched[dayKey]
		if !exists || len(day.Events) == 0 {
			continue
		}

		var cells [][]Cell
		cells = append(cells, []Cell{{Text: "Пара"}, {Text: "Занятие"}, {Text: "Аудитория"}})

		for _, ev := range day.Events {
			cells = append(cells, []Cell{
				{Text: ev.Time},
				{Text: ev.Subject},
				{Text: ev.Location},
			})
		}

		// Оборачиваем таблицу в выпадающий список.
		blocks = append(blocks, Block{
			Type:    "details",
			Summary: fmt.Sprintf("%s · %d пар", day.Name, len(day.Events)),
			Blocks: []Block{
				{
					Type:       "table",
					IsBordered: true,
					IsStriped:  true,
					Cells:      cells,
				},
			},
		})
	}

	return blocks
}

func SendWeekBlocks(bot *tgbotapi.BotAPI, chatID int64, blocks []Block) error {
	payload := RichMessageBlocksRequest{
		ChatID:      chatID,
		RichMessage: RichMessageBlocks{Blocks: blocks},
	}
	return sendRichPayload(bot, payload)
}

// ==========================================
// Общая функция отправки
// ==========================================

func sendRichPayload(bot *tgbotapi.BotAPI, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendRichMessage", bot.Token)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("ошибка API Telegram (%d): %v", resp.StatusCode, errResp)
	}

	return nil
}
