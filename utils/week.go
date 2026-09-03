package utils

import "time"

func isEvenWeek() bool {
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	now := time.Now()
	daysPassed := int(now.Sub(startDate).Hours() / 24)
	academicWeek := (daysPassed / 7) + 1
	return academicWeek%2 == 0
}

func WeekLabel() string {
	if isEvenWeek() {
		return "чётная"
	}
	return "нечётная"
}
