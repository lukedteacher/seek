package shareddto

import (
	"strings"
	"time"
)

type DayView struct {
	Day bool
}

type DaysView []DayView

func BitmaskToInitial(mask int64) string {
	var b strings.Builder

	labels := map[time.Weekday]string{
		time.Monday:    "M",
		time.Tuesday:   "T",
		time.Wednesday: "W",
		time.Thursday:  "H",
		time.Friday:    "F",
	}

	for d := time.Monday; d <= time.Friday; d++ {
		// Monday=1 -> shift 0 => 1, Tuesday=2 -> shift 1 => 2, etc.
		if mask&(1<<uint(d-1)) != 0 {
			b.WriteString(labels[d])
		}
	}
	return b.String()
}
