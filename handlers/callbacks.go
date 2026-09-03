package handlers

import (
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/schedule"
)

// HandleCallback — главный роутер callback-запросов
func HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID

	defer func() {
		bot.Send(tgbotapi.NewCallback(cb.ID, ""))
	}()

	switch {
	// === Обработка настроек уведомлений (если кнопки еще остались в меню) ===
	case data == "notify_today_morning":
		config.UpdateScheduleTime(config.ModeCustomToday, "08:00")
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Настройка сохранена: расписание будет приходить сегодня в 08:00"))

	case data == "notify_yesterday_evening":
		config.UpdateScheduleTime(config.ModeCustomYesterday, "20:00")
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Настройка сохранена: расписание будет приходить накануне в 20:00"))

	case data == "notify_disabled":
		config.UpdateScheduleTime(config.ModeDisabled, "")
		bot.Send(tgbotapi.NewMessage(chatID, "🔕 Уведомления о расписании выключены"))

	// === Существующая логика расписания ===
	case data == "today":
		sendDayByOffset(bot, chatID, 0)
	case data == "tomorrow":
		sendDayByOffset(bot, chatID, 1)
	case data == "week":
		sendWeekMenu(bot, chatID)
	case strings.HasPrefix(data, "day_"):
		dayKey := strings.TrimPrefix(data, "day_")
		sendDayWithNav(bot, chatID, dayKey)
	}
}

// sendDayByOffset — для кнопок "Сегодня" и "Завтра"
func sendDayByOffset(bot *tgbotapi.BotAPI, chatID int64, offsetDays int) {
	targetDate := time.Now().AddDate(0, 0, offsetDays)
	dayKey := weekdayToKey(targetDate.Weekday())

	_, targetWeek := targetDate.ISOWeek()
	_, currentWeek := time.Now().ISOWeek()
	weekDiff := targetWeek - currentWeek
	
	// Используем академическую четность из пакета schedule (если она экспортирована) 
	// или дублируем логику, если isEvenWeek не экспортирована. 
	// Для надежности используем проверку через schedule.GetScheduleForWeek, передавая расчет.
	isEven := isEvenAcademicWeek()
	if weekDiff%2 != 0 {
		isEven = !isEven
	}

	sched := schedule.GetScheduleForWeek(isEven)
	day, ok := sched[dayKey]
	if !ok || len(day.Events) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "🎉 В этот день пар нет!"))
		return
	}

	label := ""
	if offsetDays == 0 {
		label = "сегодня"
	} else if offsetDays == 1 {
		label = "завтра"
	}

	markdown := schedule.BuildDayMarkdown(day, label)
	if err := schedule.SendDayMarkdown(bot, chatID, markdown); err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}

// sendWeekMenu — для кнопки "Неделя"
func sendWeekMenu(bot *tgbotapi.BotAPI, chatID int64) {
	blocks := schedule.BuildWeekBlocks()
	if err := schedule.SendWeekBlocks(bot, chatID, blocks); err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}

// sendDayWithNav — для кнопок навигации ◀️ ▶️
func sendDayWithNav(bot *tgbotapi.BotAPI, chatID int64, dayKey string) {
	sched := schedule.GetSchedule()
	day, ok := sched[dayKey]
	if !ok || len(day.Events) == 0 {
		return
	}

	markdown := schedule.BuildDayMarkdown(day, "")
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

// --- Вспомогательные функции ---

func isEvenAcademicWeek() bool {
	// Та же якорная дата, что и в schedule/data.go
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	daysPassed := int(time.Since(startDate).Hours() / 24)
	academicWeek := (daysPassed / 7) + 1
	return academicWeek%2 == 0
}

func weekdayToKey(wd time.Weekday) string {
	switch wd {
	case time.Monday: return "monday"
	case time.Tuesday: return "tuesday"
	case time.Wednesday: return "wednesday"
	case time.Thursday: return "thursday"
	case time.Friday: return "friday"
	case time.Saturday: return "saturday"
	default: return "sunday"
	}
}

func dayLabel(key string) string {
	labels := map[string]string{
		"monday": "Пн", "tuesday": "Вт", "wednesday": "Ср",
		"thursday": "Чт", "friday": "Пт", "saturday": "Сб",
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
	order := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	for i, d := range order {
		if d == current && i > 0 {
			return order[i-1]
		}
	}
	return ""
}