package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/work"
)

type RichMessageMarkdownRequest struct {
	ChatID      int64               `json:"chat_id"`
	RichMessage RichMessageMarkdown `json:"rich_message"`
}

type RichMessageMarkdown struct {
	Markdown string `json:"markdown"`
}

type RichMessageBlocksRequest struct {
	ChatID      int64             `json:"chat_id"`
	RichMessage RichMessageBlocks `json:"rich_message"`
}

type RichMessageBlocks struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type       string   `json:"type"`
	Summary    string   `json:"summary,omitempty"`
	Blocks     []Block  `json:"blocks,omitempty"`
	Cells      [][]Cell `json:"cells,omitempty"`
	IsBordered bool     `json:"is_bordered,omitempty"`
	IsStriped  bool     `json:"is_striped,omitempty"`
	Text       string   `json:"text,omitempty"`
}

type Cell struct {
	Text string `json:"text"`
}

func BuildDayMarkdown(dayKey string, day DaySchedule, label string) string {
	title := fmt.Sprintf("### 📆 %s", day.Name)
	if label != "" {
		title = fmt.Sprintf("### 📆 %s (%s)", day.Name, label)
	}

	classRows := day.Events
	workRows := work.GetShiftsForDay(dayKey)

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")

	if len(classRows) > 0 {
		sb.WriteString("**Пары:**\n")
		sb.WriteString("| Время | Предмет | Место |\n")
		sb.WriteString("|-------|---------|-------|\n")
		for _, ev := range classRows {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ev.Time, ev.Subject, ev.Location))
		}
		sb.WriteString("\n")
	}

	if len(workRows) > 0 {
		sb.WriteString("**💼 Работа:**\n")
		sb.WriteString("| Время | Место |\n")
		sb.WriteString("|-------|-------|\n")
		for _, shift := range workRows {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", shift.Time, shift.Location))
		}
	}

	if len(classRows) == 0 && len(workRows) == 0 {
		sb.WriteString("Нет занятий и смен.")
	}

	return strings.TrimSpace(sb.String())
}

func BuildWeekBlocks(isEven bool, label string) []Block {
	sched := GetScheduleForWeek(isEven)
	blocks := []Block{{Type: "paragraph", Text: label}}

	for _, dayKey := range DayOrder {
		day, ok := sched[dayKey]
		if !ok {
			continue
		}
		workRows := work.GetShiftsForDay(dayKey)
		if len(day.Events) == 0 && len(workRows) == 0 {
			continue
		}

		dayBlocks := []Block{}
		if len(day.Events) > 0 {
			classCells := [][]Cell{{{Text: "Время"}, {Text: "Предмет"}, {Text: "Место"}}}
			for _, ev := range day.Events {
				classCells = append(classCells, []Cell{
					{Text: ev.Time},
					{Text: ev.Subject},
					{Text: ev.Location},
				})
			}
			dayBlocks = append(dayBlocks, Block{
				Type:       "table",
				Cells:      classCells,
				IsBordered: true,
				IsStriped:  true,
			})
		}

		if len(workRows) > 0 {
			workCells := [][]Cell{{{Text: "Время"}, {Text: "Место"}}}
			for _, shift := range workRows {
				workCells = append(workCells, []Cell{
					{Text: shift.Time},
					{Text: shift.Location},
				})
			}
			dayBlocks = append(dayBlocks, Block{
				Type:       "table",
				Cells:      workCells,
				IsBordered: true,
				IsStriped:  true,
			})
		}

		blocks = append(blocks, Block{
			Type:    "details",
			Summary: day.Name,
			Blocks:  dayBlocks,
		})
	}

	return blocks
}

func SendDayMarkdown(bot *tgbotapi.BotAPI, chatID int64, markdown string) error {
	payload := RichMessageMarkdownRequest{
		ChatID:      chatID,
		RichMessage: RichMessageMarkdown{Markdown: markdown},
	}
	return sendRichPayload(bot, payload)
}

func SendWeekBlocks(bot *tgbotapi.BotAPI, chatID int64, blocks []Block) error {
	payload := RichMessageBlocksRequest{
		ChatID:      chatID,
		RichMessage: RichMessageBlocks{Blocks: blocks},
	}
	return sendRichPayload(bot, payload)
}

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
