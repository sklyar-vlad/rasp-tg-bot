package schedule

import "time"

// Event представляет одну пару или окно
type Event struct {
	Time     string // например: "1 · 11:40" или "-"
	Subject  string // например: "пр. · Экономика"
	Location string // например: "9-213 · Володин"
	IsWindow bool   // если это окно, форматируем иначе
}

type DaySchedule struct {
	Name   string  // "Понедельник"
	Events []Event // Список событий дня
}

var oddWeek = map[string]DaySchedule{
	"monday": {
		Name: "Понедельник",
		Events: []Event{
			{Time: "3 · 11:40", Subject: "пр. · Экономика", Location: "9-213 · Володин"},
			{Time: "4 · 13:45", Subject: "Физкультура (электив)", Location: "-"},
		},
	},
	"tuesday": {
		Name: "Вторник",
		Events: []Event{
			{Time: "1 · 8:00", Subject: "лб · ООЯ и СП 2", Location: "7а-206а · Горюнов Ю.Ю."},
			{Time: "2 · 9:50", Subject: "лб · Я и МП 2", Location: "7а-512 · Барсукова О.Ю."},
		},
	},
	"wednesday": {
		Name: "Среда",
		Events: []Event{
			{Time: "4 · 13:45", Subject: "лек. · ФМП и ТО", Location: "7а-508 · Пичугина П.Г."},
			{Time: "5 · 15:35", Subject: "пр. · БЖД", Location: "8-301 · Авдонина Л.А."},
		},
	},
	"thursday": {
		Name: "Четверг",
		Events: []Event{
			{Time: "1 · 8:00", Subject: "пр. · ФМП и ТО", Location: "7а-508 · Пичугина П.Г."},
			{Time: "2 · 9:50", Subject: "Физкультура (электив)", Location: "-"},
			{Time: "3 · 11:40", Subject: "лек. · Обыкн. и ДУ", Location: "5-102 · Валовик Д.В."},
			{Time: "4 · 13:45", Subject: "пр. · Философия", Location: "76-205 · Кириллов Г.М."},
		},
	},
	"friday": {
		Name: "Пятница",
		Events: []Event{
			{Time: "1 · 8:00", Subject: "пр. · Функции многих переменных и ТО", Location: "7а-508 · Пичугина П.Г."},
			{Time: "2 · 9:50", Subject: "лек. · БЖД", Location: "76-207 · Полянцева Е.А."},
			{Time: "3 · 11:40", Subject: "лек. · Экономика", Location: "76-207 · Володин В.М."},
		},
	},
	"saturday": {
		Name: "Суббота",
		Events: []Event{
			{Time: "3 · 11:40", Subject: "пр. · Обыкн. ДУ", Location: "3-211 · Валовик Д.В."},
			{Time: "4 · 13:45", Subject: "лаб. · Иностранный язык", Location: "8-802а/б · Данкова / Мусорина"},
		},
	},
}

var evenWeek = map[string]DaySchedule{
	"monday": {
		Name: "Понедельник",
		Events: []Event{
			{Time: "2 · 9:50", Subject: "Кураторский час", Location: "76-202"},
			{Time: "3 · 11:40", Subject: "пр. · ОПД", Location: "7а-305а · Катышева М.А."},
			{Time: "4 · 13:45", Subject: "Физкультура (электив)", Location: "-"},
		},
	},
	"tuesday": {
		Name: "Вторник",
		Events: []Event{
			{Time: "1 · 8:00", Subject: "л6 · ООЯ и СП 2", Location: "7а-206а · Горюнов Ю.Ю."},
			{Time: "2 · 9:50", Subject: "л6 · Я и МП 2", Location: "7а-512 · Барсукова О.Ю."},
		},
	},
	"wednesday": {
		Name: "Среда",
		Events: []Event{
			{Time: "3 · 13:45", Subject: "лек. · ФМП и ТО", Location: "7а-508 · Пичугина П.Г."},
			{Time: "5 · 15:35", Subject: "лек. · ООЯ и СП", Location: "7а-512 · Горюнов Ю.Ю."},
		},
	},
	"thursday": {
		Name: "Четверг",
		Events: []Event{
			{Time: "2 · 9:50", Subject: "Физкультура (электив)", Location: "-"},
			{Time: "3 · 11:40", Subject: "лек. · Философия", Location: "76-207 · Кириллов Г.М."},
			{Time: "4 · 13:45", Subject: "пр. · Философия", Location: "76-205 · Кириллов Г.М."},
		},
	},
	"friday": {
		Name: "Пятница",
		Events: []Event{
			{Time: "1 · 8:00", Subject: "пр. · Функции многих переменных и ТО", Location: "7а-508 · Пичугина П.Г."},
			{Time: "2 · 9:50", Subject: "лек. · БЖД", Location: "76-207 · Полянцева Е.А."},
			{Time: "3 · 11:40", Subject: "лек. · Основы ПД", Location: "76-207 · Кирюхин Ю.Г."},
		},
	},
	"saturday": {
		Name: "Суббота",
		Events: []Event{
			{Time: "3 · 11:40", Subject: "пр. · Обыкн. ДУ", Location: "3-211 · Валовик Д.В."},
			{Time: "4 · 13:45", Subject: "лаб. · Иностранный язык", Location: "8-802а/б · Данкова / Мусорина"},
		},
	},
}

func GetSchedule() map[string]DaySchedule {
	if isEvenWeek() {
		return evenWeek
	}
	return oddWeek
}

func GetScheduleForWeek(isEven bool) map[string]DaySchedule {
	if isEven {
		return evenWeek
	}
	return oddWeek
}

var DayOrder = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

// isEvenWeek определяет четность учебной недели относительно якорной даты.
// Якорная дата: 1 сентября 2024 года считается 1-й (нечетной) неделей.
// Этот метод гарантирует, что 1 сентября любого года всегда будет нечетной неделей.
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