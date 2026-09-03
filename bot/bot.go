package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/handlers"
)

func Start(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if !isAllowed(update) {
			handlers.HandleMessageFromGuest(bot, update.Message)
			continue
		}

		if update.Message != nil {
			handlers.HandleMessage(bot, update.Message)
			handlers.HandleAction(bot, update.Message)

			switch update.Message.Text {
			case "/start":
				handlers.HandleStart(bot, update.Message)
			case "/help":
				handlers.HandleHelp(bot, update.Message)
			case "/time":
				handlers.HandleSettings(bot, update.Message)
			default:
				if strings.HasPrefix(update.Message.Text, "/time ") {
					args := strings.TrimPrefix(update.Message.Text, "/time ")
					handlers.HandleSetTime(bot, update.Message, args)
				}
			}
		}

		if update.CallbackQuery != nil {
			handlers.HandleCallback(bot, update.CallbackQuery)
		}
	}
}

func isAllowed(update tgbotapi.Update) bool {
	var userID int64
	if update.Message != nil {
		userID = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
	} else {
		return false
	}

	for _, id := range config.AllowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}
