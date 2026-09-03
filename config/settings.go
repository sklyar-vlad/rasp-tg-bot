package config

import (
	"encoding/json"
	"os"
	"sync"
)

type NotifyMode string

const (
	ModeDisabled        NotifyMode = "disabled"
	ModeCustomToday     NotifyMode = "custom_today"     // Произвольное время сегодня
	ModeCustomYesterday NotifyMode = "custom_yesterday" // Произвольное время накануне
)

type Settings struct {
	NotifyMode       NotifyMode `json:"notify_mode"`
	CustomTime       string     `json:"custom_time"` // Формат "ЧЧ:ММ", например "08:30"
	DailyMessageTime string     `json:"daily_message_time"`
	DailyMessages    []string   `json:"daily_messages"`
}

const settingsFile = "settings.json"

var (
	currentSettings Settings
	settingsMu      sync.RWMutex
)

func LoadSettings() Settings {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		currentSettings = Settings{
			NotifyMode:       ModeCustomToday,
			CustomTime:       "08:00", // Значение по умолчанию
			DailyMessageTime: "33 03 * * *",
			DailyMessages: []string{
				"☀️ Доброе утро! Отличного дня!",
				"💪 Ты справишься со всем сегодня!",
			},
		}
		SaveSettings()
		return currentSettings
	}

	json.Unmarshal(data, &currentSettings)
	// Защита от пустого времени при первом запуске старого файла
	if currentSettings.CustomTime == "" {
		currentSettings.CustomTime = "08:00"
		SaveSettings()
	}
	return currentSettings
}

func SaveSettings() {
	data, _ := json.MarshalIndent(currentSettings, "", "  ")
	os.WriteFile(settingsFile, data, 0644)
}

func GetSettings() Settings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return currentSettings
}

// UpdateScheduleTime обновляет режим и время
func UpdateScheduleTime(mode NotifyMode, timeStr string) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	currentSettings.NotifyMode = mode
	currentSettings.CustomTime = timeStr
	SaveSettings()
}
