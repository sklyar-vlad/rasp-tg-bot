package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/admin"
	"github.com/sklyar-vlad/rasp-tg-bot/config"
)

func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.From.ID == config.MainUserID && msg.Text != "" {
		admin.LogMessage(msg.From.UserName, msg.Text)
	}
}

func HandleMessageFromGuest(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Отказано в доступе"))
}
