package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Загружаем .env при инициализации пакета
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Файл .env не найден. Убедитесь, что переменная BOT_TOKEN задана в системе.")
	}
}

var AllowedUsers = []int64{
	5781599483,
	1436576269,
}

// BotToken теперь читается из переменной окружения, а не зашит в код
var BotToken = os.Getenv("BOT_TOKEN")

const MainUserID int64 = 5781599483