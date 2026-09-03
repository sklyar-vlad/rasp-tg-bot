package admin

import (
	"fmt"
	"os"
	"time"
)

const LogFile = "user_messages.log"

func LogMessage(username, text string) {
	f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] @%s: %s\n", ts, username, text)
}
