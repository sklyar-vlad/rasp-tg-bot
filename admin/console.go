package admin

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sklyar-vlad/rasp-tg-bot/config"
)

func StartConsoleListener(bot *tgbotapi.BotAPI) {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("🖥 Консоль администратора активна.")
		fmt.Println("Введите текст → Enter = отправка пользователю.")
		fmt.Println("Команды: /users — список разрешённых, exit — выход.")

		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			if text == "exit" {
				fmt.Println("Остановка бота...")
				os.Exit(0)
			}
			if text == "/users" {
				fmt.Printf("Whitelist: %v\n", config.AllowedUsers)
				continue
			}

			msg := tgbotapi.NewMessage(config.MainUserID, text)
			if _, err := bot.Send(msg); err != nil {
				fmt.Printf("❌ Ошибка отправки: %v\n", err)
			} else {
				fmt.Println("✅ Отправлено.")
			}
		}

		// ✅ Правильная проверка ошибок после завершения цикла Scan
		if err := scanner.Err(); err != nil {
			fmt.Printf("❌ Ошибка чтения из консоли: %v\n", err)
		}
	}()
}
