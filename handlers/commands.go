package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/notifications"
	"github.com/sklyar-vlad/rasp-tg-bot/utils"
)

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Сегодня"),
			tgbotapi.NewKeyboardButton("📆 Завтра"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗓 Неделя"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚙️ Настройки"),
		),
	)
	keyboard.ResizeKeyboard = true

	text := fmt.Sprintf(
		"👋 Привет! Я бот-помощник.\n\n"+
			"📌 Текущая неделя: *%s*\n\n"+
			"Выбери, что посмотреть:",
		utils.WeekLabel(),
	)

	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ParseMode = "Markdown"
	m.ReplyMarkup = keyboard
	bot.Send(m)
}

func HandleSettings(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	settings := config.GetSettings()

	modeLabel := "🔕 Выключено"
	switch settings.NotifyMode {
	case config.ModeCustomToday:
		modeLabel = fmt.Sprintf("🌅 Сегодня в %s", settings.CustomTime)
	case config.ModeCustomYesterday:
		modeLabel = fmt.Sprintf("🌙 Накануне в %s", settings.CustomTime)
	}

	text := fmt.Sprintf(
		"⚙️ *Настройки уведомлений*\n\n"+
			"📬 Расписание: %s\n\n"+
			"🕒 *Как изменить время?*\n"+
			"Отправьте команду:\n"+
			"• `/time 08:30` — присылать сегодня в 08:30\n"+
			"• `/time вчера 19:00` — присылать накануне в 19:00\n"+
			"• `/time off` — выключить уведомления",
		modeLabel,
	)

	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ParseMode = "Markdown"
	bot.Send(m)
}

func HandleSetTime(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, args string) {
	args = strings.TrimSpace(args)

	if args == "off" || args == "выкл" {
		config.UpdateScheduleTime(config.ModeDisabled, "")
		notifications.ReloadScheduler(bot)
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "🔕 Уведомления о расписании выключены."))
		return
	}

	parts := strings.Fields(args)
	var mode config.NotifyMode = config.ModeCustomToday
	var timeStr string

	if len(parts) == 2 && (parts[0] == "вчера" || parts[0] == "накануне") {
		mode = config.ModeCustomYesterday
		timeStr = parts[1]
	} else if len(parts) == 1 {
		timeStr = parts[0]
	} else {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный формат. Используйте:\n`/time 08:30` или `/time вчера 19:00`"))
		return
	}

	if !isValidTimeFormat(timeStr) {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Неверный формат времени. Используйте ЧЧ:ММ (например, 08:30 или 19:00)."))
		return
	}

	config.UpdateScheduleTime(mode, timeStr)
	notifications.ReloadScheduler(bot)

	dayWord := "сегодня"
	if mode == config.ModeCustomYesterday {
		dayWord = "накануне"
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("✅ Сохранено! Расписание будет приходить %s в %s.", dayWord, timeStr)))
}

func isValidTimeFormat(t string) bool {
	if len(t) != 5 || t[2] != ':' {
		return false
	}
	hh := t[0:2]
	mm := t[3:5]
	return isDigits(hh) && isDigits(mm)
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func HandleAction(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Text {
	case "📅 Сегодня":
		sendDayByOffset(bot, msg.Chat.ID, 0)
	case "📆 Завтра":
		sendDayByOffset(bot, msg.Chat.ID, 1)
	case "🗓 Неделя":
		sendWeekMenu(bot, msg.Chat.ID)
	case "⚙️ Настройки":
		HandleSettings(bot, msg)
	}
}

func HandleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "Используй /start для вызова меню.\nИспользуй /time для настройки уведомлений."
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, text))
}
