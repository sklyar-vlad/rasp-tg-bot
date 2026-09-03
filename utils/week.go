package utils

import "time"

func isEvenWeek() bool {
	// Якорная дата: 1 сентября 2024 года, 00:00
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	now := time.Now()

	// Считаем количество полных дней, прошедших с якорной даты
	daysPassed := int(now.Sub(startDate).Hours() / 24)

	// Вычисляем номер академической недели (начиная с 1)
	academicWeek := (daysPassed / 7) + 1

	// Если номер недели делится на 2 без остатка -> она четная
	// Если остаток 1 -> она нечетная
	return academicWeek%2 == 0
}

func WeekLabel() string {
	if isEvenWeek() {
		return "чётная"
	}
	return "нечётная"
}