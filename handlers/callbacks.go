package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sklyar-vlad/rasp-tg-bot/schedule"
	"github.com/sklyar-vlad/rasp-tg-bot/work"
)

func HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID

	defer func() {
		bot.Send(tgbotapi.NewCallback(cb.ID, ""))
	}()

	switch {
	case data == "today":
		sendDayByOffset(bot, chatID, 0)
	case data == "tomorrow":
		sendDayByOffset(bot, chatID, 1)
	case data == "week":
		sendWeekMenu(bot, chatID)
	case data == "week_even":
		sendWeekByParity(bot, chatID, true)
	case data == "week_odd":
		sendWeekByParity(bot, chatID, false)
	case strings.HasPrefix(data, "day_"):
		dayKey := strings.TrimPrefix(data, "day_")
		sendDayWithNav(bot, chatID, dayKey)
	case strings.HasPrefix(data, "work_day_"):
		handleWorkDayToggle(bot, cb, strings.TrimPrefix(data, "work_day_"))
	case data == "work_done":
		handleWorkDone(bot, chatID)
	case data == "work_cancel":
		delete(shiftStates, chatID)
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Настройка смен отменена"))
	case data == "work_add_more":
		startWorkShiftSetup(bot, chatID)
	case data == "work_reset_all":
		work.ResetShifts()
		m := tgbotapi.NewMessage(chatID, "🗑 Все смены удалены")
		m.ParseMode = "Markdown"
		keyboard := buildWorkPostSaveKeyboard()
		m.ReplyMarkup = keyboard
		bot.Send(m)
	case data == "work_finish":
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Настройка работы завершена"))
	}
}

func sendDayByOffset(bot *tgbotapi.BotAPI, chatID int64, offsetDays int) {
	targetDate := time.Now().AddDate(0, 0, offsetDays)
	dayKey := weekdayToKey(targetDate.Weekday())

	_, targetWeek := targetDate.ISOWeek()
	_, currentWeek := time.Now().ISOWeek()
	weekDiff := targetWeek - currentWeek

	isEven := isEvenAcademicWeek()
	if weekDiff%2 != 0 {
		isEven = !isEven
	}

	sched := schedule.GetScheduleForWeek(isEven)
	day, ok := sched[dayKey]
	if !ok {
		day = schedule.DaySchedule{Name: dayNameByKey(dayKey)}
	}
	if len(day.Events) == 0 && len(work.GetShiftsForDay(dayKey)) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "🎉 В этот день пар нет!"))
		return
	}

	label := ""
	if offsetDays == 0 {
		label = "сегодня"
	} else if offsetDays == 1 {
		label = "завтра"
	}

	markdown := schedule.BuildDayMarkdown(dayKey, day, label)
	if err := schedule.SendDayMarkdown(bot, chatID, markdown); err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}

func sendWeekMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Чётная неделя", "week_even"),
			tgbotapi.NewInlineKeyboardButtonData("Нечётная неделя", "week_odd"),
		),
	)
	m := tgbotapi.NewMessage(chatID, "Выберите тип недели:")
	m.ParseMode = "Markdown"
	m.ReplyMarkup = keyboard
	bot.Send(m)
}

func sendWeekByParity(bot *tgbotapi.BotAPI, chatID int64, isEven bool) {
	label := "Нечётная неделя"
	if isEven {
		label = "Чётная неделя"
	}
	markdown := schedule.BuildWeekMarkdown(isEven, label)
	m := tgbotapi.NewMessage(chatID, markdown)
	m.ParseMode = "Markdown"
	bot.Send(m)
}

func sendDayWithNav(bot *tgbotapi.BotAPI, chatID int64, dayKey string) {
	sched := schedule.GetSchedule()
	day, ok := sched[dayKey]
	if !ok {
		day = schedule.DaySchedule{Name: dayNameByKey(dayKey)}
	}
	if len(day.Events) == 0 && len(work.GetShiftsForDay(dayKey)) == 0 {
		return
	}

	markdown := schedule.BuildDayMarkdown(dayKey, day, "")
	if err := schedule.SendDayMarkdown(bot, chatID, markdown); err != nil {
		log.Printf("Ошибка отправки: %v", err)
		return
	}

	prevKey := prevDayKey(dayKey)
	nextKey := nextDayKey(dayKey)

	var row []tgbotapi.InlineKeyboardButton
	if prevKey != "" {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️ "+dayLabel(prevKey), "day_"+prevKey))
	}
	if nextKey != "" {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(dayLabel(nextKey)+" ▶️", "day_"+nextKey))
	}

	if len(row) > 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
		m := tgbotapi.NewMessage(chatID, " ")
		m.ReplyMarkup = keyboard
		bot.Send(m)
	}
}

func isEvenAcademicWeek() bool {
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	daysPassed := int(time.Since(startDate).Hours() / 24)
	academicWeek := (daysPassed / 7) + 1
	return academicWeek%2 == 0
}

func weekdayToKey(wd time.Weekday) string {
	switch wd {
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	default:
		return "sunday"
	}
}

func dayLabel(key string) string {
	labels := map[string]string{
		"monday": "Пн", "tuesday": "Вт", "wednesday": "Ср",
		"thursday": "Чт", "friday": "Пт", "saturday": "Сб", "sunday": "Вс",
	}
	return labels[key]
}

func nextDayKey(current string) string {
	order := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	for i, d := range order {
		if d == current && i < len(order)-1 {
			return order[i+1]
		}
	}
	return ""
}

func prevDayKey(current string) string {
	order := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for i, d := range order {
		if d == current && i > 0 {
			return order[i-1]
		}
	}
	return ""
}

func dayNameByKey(dayKey string) string {
	names := map[string]string{
		"monday": "Понедельник", "tuesday": "Вторник", "wednesday": "Среда",
		"thursday": "Четверг", "friday": "Пятница", "saturday": "Суббота", "sunday": "Воскресенье",
	}
	return names[dayKey]
}

func handleWorkDayToggle(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, dayKey string) {
	state, ok := shiftStates[cb.Message.Chat.ID]
	if !ok {
		state = &ShiftState{Step: 0, Days: []string{}}
		shiftStates[cb.Message.Chat.ID] = state
	}
	if state.Step != 0 {
		return
	}

	state.Days = toggleSelectedDay(state.Days, dayKey)
	text := fmt.Sprintf("**Настройка смены**\n\nВыберите дни:\n%s", formatSelectedDays(state.Days))
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text)
	edit.ParseMode = "Markdown"
	keyboard := buildWorkDaysKeyboard(state.Days)
	edit.ReplyMarkup = &keyboard
	bot.Send(edit)
}

func handleWorkDone(bot *tgbotapi.BotAPI, chatID int64) {
	state, ok := shiftStates[chatID]
	if !ok || len(state.Days) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Выберите хотя бы один день"))
		return
	}
	state.Step = 1

	m := tgbotapi.NewMessage(chatID, "Введите время (09:00-18:00)")
	m.ParseMode = "Markdown"
	bot.Send(m)
}
