package utils

import "time"

func IsEvenWeek() bool {
	_, week := time.Now().ISOWeek()
	return week%2 == 0
}

func WeekLabel() string {
	if IsEvenWeek() {
		return "чётная"
	}
	return "нечётная"
}