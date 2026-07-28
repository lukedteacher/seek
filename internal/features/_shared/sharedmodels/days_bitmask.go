package sharedmodels

import (
	"encoding/json"
	"strings"
	"time"
)

type DaysBitmask int

const (
	BitMonday = 1 << iota
	BitTuesday
	BitWednesday
	BitThursday
	BitFriday
)

func (m DaysBitmask) IsDaySet(day Day) bool {
	return int(m)&day.Bit() != 0
}

func (m DaysBitmask) String() string {
	var b strings.Builder
	labels := map[time.Weekday]string{
		time.Monday:    "M",
		time.Tuesday:   "T",
		time.Wednesday: "W",
		time.Thursday:  "H",
		time.Friday:    "F",
	}
	for d := time.Monday; d <= time.Friday; d++ {
		if int(m)&(1<<uint(d-1)) != 0 {
			b.WriteString(labels[d])
		}
	}
	return b.String()
}

func (m DaysBitmask) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(m))
}

func (m *DaysBitmask) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*m = DaysBitmask(i)
	return nil
}
