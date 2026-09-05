package work

import "sync"

type Shift struct {
	Time     string
	Location string
}

var (
	mu     sync.RWMutex
	shifts = map[string][]Shift{}
)

func AddShift(dayKey, timeRange, location string) {
	mu.Lock()
	defer mu.Unlock()
	shifts[dayKey] = append(shifts[dayKey], Shift{
		Time:     timeRange,
		Location: location,
	})
}

func GetShiftsForDay(dayKey string) []Shift {
	mu.RLock()
	defer mu.RUnlock()
	items := shifts[dayKey]
	out := make([]Shift, len(items))
	copy(out, items)
	return out
}

func ResetShifts() {
	mu.Lock()
	defer mu.Unlock()
	shifts = map[string][]Shift{}
}
