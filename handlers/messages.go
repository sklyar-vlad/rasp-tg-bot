package handlers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/admin"
	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/work"
)

type ShiftState struct {
	Step     int
	Days     []string
	Time     string
	Location string
}

var shiftStates = map[int64]*ShiftState{}

var shiftDayKeys = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

var shiftDayLabels = map[string]string{
	"monday": "Пн", "tuesday": "Вт", "wednesday": "Ср", "thursday": "Чт",
	"friday": "Пт", "saturday": "Сб", "sunday": "Вс",
}

var shiftTimeRe = regexp.MustCompile(`^\d{2}:\d{2}-\d{2}:\d{2}$`)

func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.From.ID == config.MainUserID && msg.Text != "" {
		admin.LogMessage(msg.From.UserName, msg.Text)
	}

	if msg.Text == "" || strings.HasPrefix(msg.Text, "/") {
		return
	}

	state, ok := shiftStates[msg.Chat.ID]
	if !ok {
		return
	}

	switch state.Step {
	case 1:
		value := strings.TrimSpace(msg.Text)
		if !shiftTimeRe.MatchString(value) {
			m := tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный формат. Введите время в формате `09:00-18:00`")
			m.ParseMode = "Markdown"
			bot.Send(m)
			return
		}
		state.Time = value
		state.Step = 2

		m := tgbotapi.NewMessage(msg.Chat.ID, "Введите место")
		m.ParseMode = "Markdown"
		bot.Send(m)
	case 2:
		value := strings.TrimSpace(msg.Text)
		if value == "" {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Место не может быть пустым"))
			return
		}

		state.Location = value
		for _, day := range state.Days {
			work.AddShift(day, state.Time, state.Location)
		}

		confirmation := fmt.Sprintf(
			"✅ Смена сохранена\n\nДни: %s\nВремя: %s\nМесто: %s",
			formatSelectedDays(state.Days),
			state.Time,
			state.Location,
		)
		m := tgbotapi.NewMessage(msg.Chat.ID, confirmation)
		m.ParseMode = "Markdown"
		keyboard := buildWorkPostSaveKeyboard()
		m.ReplyMarkup = keyboard
		bot.Send(m)

		delete(shiftStates, msg.Chat.ID)
	}
}

func HandleMessageFromGuest(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Отказано в доступе"))
}

func startWorkShiftSetup(bot *tgbotapi.BotAPI, chatID int64) {
	shiftStates[chatID] = &ShiftState{Step: 0, Days: []string{}}
	text := "**Настройка смены**\n\nВыберите дни:"
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	keyboard := buildWorkDaysKeyboard(nil)
	m.ReplyMarkup = keyboard
	bot.Send(m)
}

func buildWorkDaysKeyboard(selected []string) tgbotapi.InlineKeyboardMarkup {
	selectedMap := map[string]bool{}
	for _, day := range selected {
		selectedMap[day] = true
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		buildWorkDayButton("monday", selectedMap),
		buildWorkDayButton("tuesday", selectedMap),
		buildWorkDayButton("wednesday", selectedMap),
		buildWorkDayButton("thursday", selectedMap),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		buildWorkDayButton("friday", selectedMap),
		buildWorkDayButton("saturday", selectedMap),
		buildWorkDayButton("sunday", selectedMap),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Готово", "work_done"),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "work_cancel"),
	})
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildWorkDayButton(dayKey string, selected map[string]bool) tgbotapi.InlineKeyboardButton {
	label := shiftDayLabels[dayKey]
	if selected[dayKey] {
		label = "✓ " + label
	}
	return tgbotapi.NewInlineKeyboardButtonData(label, "work_day_"+dayKey)
}

func buildWorkPostSaveKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить ещё", "work_add_more"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Сбросить все", "work_reset_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "work_finish"),
		),
	)
}

func toggleSelectedDay(days []string, dayKey string) []string {
	for i, day := range days {
		if day == dayKey {
			return append(days[:i], days[i+1:]...)
		}
	}
	return append(days, dayKey)
}

func formatSelectedDays(days []string) string {
	if len(days) == 0 {
		return "—"
	}

	ranks := map[string]int{}
	for idx, key := range shiftDayKeys {
		ranks[key] = idx
	}
	copyDays := append([]string(nil), days...)
	sort.Slice(copyDays, func(i, j int) bool {
		return ranks[copyDays[i]] < ranks[copyDays[j]]
	})

	labels := make([]string, 0, len(copyDays))
	for _, day := range copyDays {
		labels = append(labels, shiftDayLabels[day])
	}
	return strings.Join(labels, ", ")
}
