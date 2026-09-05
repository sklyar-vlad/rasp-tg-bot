package notifications

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"

	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/schedule"
)

var (
	cronInstance *cron.Cron
	cronMu       sync.Mutex
)

func StartScheduler(bot *tgbotapi.BotAPI) {
	rebuildCron(bot)
}

func rebuildCron(bot *tgbotapi.BotAPI) {
	cronMu.Lock()
	defer cronMu.Unlock()

	// 1. Останавливаем старый планировщик, если он есть
	if cronInstance != nil {
		cronInstance.Stop()
	}

	// 2. Создаем новый экземпляр
	cronInstance = cron.New()
	settings := config.GetSettings() // Берем самые свежие настройки из памяти

	// 3. Добавляем задачу для расписания
	if settings.NotifyMode != config.ModeDisabled && settings.CustomTime != "" {
		cronExpr := timeToCron(settings.CustomTime)

		var jobFunc func()
		switch settings.NotifyMode {
		case config.ModeCustomToday:
			jobFunc = func() { sendScheduleForToday(bot) }
			log.Printf("⏰ Расписание: сегодня в %s (cron: %s)", settings.CustomTime, cronExpr)
		case config.ModeCustomYesterday:
			jobFunc = func() { sendScheduleForTomorrow(bot) }
			log.Printf("⏰ Расписание: накануне в %s (cron: %s)", settings.CustomTime, cronExpr)
		}

		if jobFunc != nil {
			_, err := cronInstance.AddFunc(cronExpr, jobFunc)
			if err != nil {
				log.Printf("❌ Ошибка cron (расписание): %v", err)
			}
		}
	} else {
		log.Println("⏰ Уведомления о расписании выключены")
	}

	// 4. Добавляем задачу для ежедневных сообщений (комплименты/уведомления)
	if len(settings.DailyMessages) > 0 && settings.DailyMessageTime != "" {
		_, err := cronInstance.AddFunc(settings.DailyMessageTime, func() {
			sendDailyCustomMessage(bot)
		})
		if err != nil {
			log.Printf("❌ Ошибка cron (daily_message): %v", err)
		}
		log.Printf("⏰ Ежедневное сообщение: %s", settings.DailyMessageTime)
	}

	// 5. Запускаем новый планировщик
	cronInstance.Start()
}

func timeToCron(timeStr string) string {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		// Формат: минута час день_месяца месяц день_недели (ровно 5 полей)
		// Например, для "02:50" вернёт "50 02 * * *"
		return fmt.Sprintf("%s %s * * *", parts[1], parts[0])
	}
	// Fallback на 08:00 (тоже 5 полей)
	return "0 8 * * *"
}
func ReloadScheduler(bot *tgbotapi.BotAPI) {
	rebuildCron(bot)
}

// ✅ ДОБАВЛЕНО: отправка расписания на сегодня
func sendScheduleForToday(bot *tgbotapi.BotAPI) {
	dayKey := todayKey()
	sched := schedule.GetSchedule()
	day, ok := sched[dayKey]
	if !ok {
		return
	}
	if len(day.Events) == 0 {
		bot.Send(tgbotapi.NewMessage(config.MainUserID, "🎉 Сегодня пар нет! Отдыхай."))
		return
	}

	markdown := schedule.BuildDayMarkdown(dayKey, day, "ежедневное")
	if err := schedule.SendDayMarkdown(bot, config.MainUserID, markdown); err != nil {
		log.Printf("Ошибка отправки расписания: %v", err)
	}
}

// ✅ ДОБАВЛЕНО: отправка расписания на завтра
func sendScheduleForTomorrow(bot *tgbotapi.BotAPI) {
	targetDate := time.Now().AddDate(0, 0, 1)
	dayKey := weekdayToKey(targetDate.Weekday())

	_, targetWeek := targetDate.ISOWeek()
	_, currentWeek := time.Now().ISOWeek()
	weekDiff := targetWeek - currentWeek

	// Академическая четность
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	daysPassed := int(time.Since(startDate).Hours() / 24)
	academicWeek := (daysPassed / 7) + 1
	isEven := academicWeek%2 == 0

	if weekDiff%2 != 0 {
		isEven = !isEven
	}

	sched := schedule.GetScheduleForWeek(isEven)
	day, ok := sched[dayKey]
	if !ok || len(day.Events) == 0 {
		bot.Send(tgbotapi.NewMessage(config.MainUserID, "🎉 Завтра пар нет! Отдыхай."))
		return
	}

	markdown := schedule.BuildDayMarkdown(dayKey, day, "на завтра")
	if err := schedule.SendDayMarkdown(bot, config.MainUserID, markdown); err != nil {
		log.Printf("Ошибка отправки расписания: %v", err)
	}
}

func sendDailyCustomMessage(bot *tgbotapi.BotAPI) {
	settings := config.GetSettings()
	if len(settings.DailyMessages) == 0 {
		return
	}
	msg := settings.DailyMessages[rand.Intn(len(settings.DailyMessages))]
	bot.Send(tgbotapi.NewMessage(config.MainUserID, msg))
}

func todayKey() string {
	switch time.Now().Weekday() {
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
