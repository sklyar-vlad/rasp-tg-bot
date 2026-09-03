package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sklyar-vlad/rasp-tg-bot/admin"
	"github.com/sklyar-vlad/rasp-tg-bot/bot"
	"github.com/sklyar-vlad/rasp-tg-bot/config"
	"github.com/sklyar-vlad/rasp-tg-bot/notifications"
)

func main() {
	// Загружаем настройки (создаст файл, если его нет)
	config.LoadSettings()

	botInstance, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем всё параллельно
	admin.StartConsoleListener(botInstance)
	notifications.StartScheduler(botInstance)

	log.Println("🤖 Бот запущен. Whitelist:", config.AllowedUsers)

	bot.Start(botInstance)
}