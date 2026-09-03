package config

import (
	"os"
	"strconv"
)

var (
	admin, _    = strconv.Atoi(os.Getenv("ADMIN"))
	mainUser, _ = strconv.Atoi(os.Getenv("MAIN_USER"))
)

var AllowedUsers = []int64{
	int64(admin),
	int64(mainUser),
}

var BotToken = os.Getenv("BOT_TOKEN")

var MainUserID int64 = int64(mainUser)
