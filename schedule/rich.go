package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/work"
)

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
	IsCompact  bool     `json:"is_compact,omitempty"`
	IsStriped  bool     `json:"is_striped,omitempty"`
	Text       string   `json:"text,omitempty"`
}

type Cell struct {
	Text string `json:"text"`
}

func BuildDayBlocks(dayKey string, day DaySchedule, label string) []Block {
	title := fmt.Sprintf("📆 %s", day.Name)
	if label != "" {
		title = fmt.Sprintf("📆 %s (%s)", day.Name, label)
	}

	blocks := []Block{{Type: "paragraph", Text: title}}

	workRows := work.GetShiftsForDay(dayKey)
	if len(day.Events) > 0 {
		classCells := [][]Cell{{{Text: "Время"}, {Text: "Предмет"}, {Text: "Место"}}}
		for _, ev := range day.Events {
			classCells = append(classCells, []Cell{
				{Text: ev.Time},
				{Text: ev.Subject},
				{Text: ev.Location},
			})
		}
		blocks = append(blocks, Block{
			Type:       "table",
			Cells:      classCells,
			IsBordered: true,
			IsCompact:  true,
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
		blocks = append(blocks, Block{
			Type:       "table",
			Cells:      workCells,
			IsBordered: true,
			IsCompact:  true,
		})
	}

	return blocks
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
				IsCompact:  true,
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
				IsCompact:  true,
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

func SendDayBlocks(bot *tgbotapi.BotAPI, chatID int64, blocks []Block) error {
	payload := RichMessageBlocksRequest{
		ChatID:      chatID,
		RichMessage: RichMessageBlocks{Blocks: blocks},
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
