package schedule

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/work"
)

func BuildDayMarkdown(dayKey string, day DaySchedule, label string) string {
	title := fmt.Sprintf("📆 %s", day.Name)
	if label != "" {
		title = fmt.Sprintf("📆 %s (%s)", day.Name, label)
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

func BuildWeekMarkdown(isEven bool, label string) string {
	sched := GetScheduleForWeek(isEven)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n\n", label))

	for _, dayKey := range DayOrder {
		day, ok := sched[dayKey]
		if !ok {
			continue
		}
		if len(day.Events) == 0 && len(work.GetShiftsForDay(dayKey)) == 0 {
			continue
		}
		sb.WriteString(BuildDayMarkdown(dayKey, day, ""))
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}

func SendDayMarkdown(bot *tgbotapi.BotAPI, chatID int64, markdown string) error {
	m := tgbotapi.NewMessage(chatID, markdown)
	m.ParseMode = "Markdown"
	_, err := bot.Send(m)
	return err
}
