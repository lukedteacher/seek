package shareddto

import (
	"strings"
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type DaysFormView struct {
	Monday    bool `json:"monday"`
	Tuesday   bool `json:"tuesday"`
	Wednesday bool `json:"wednesday"`
	Thursday  bool `json:"thursday"`
	Friday    bool `json:"friday"`
}

func (ds *DaysFormView) setDay(day sharedmodels.Day, set bool) {
	switch day {
	case sharedmodels.Day(time.Monday):
		ds.Monday = set
	case sharedmodels.Day(time.Tuesday):
		ds.Tuesday = set
	case sharedmodels.Day(time.Wednesday):
		ds.Wednesday = set
	case sharedmodels.Day(time.Thursday):
		ds.Thursday = set
	case sharedmodels.Day(time.Friday):
		ds.Friday = set
	}
}

func (ds DaysFormView) IsDaySet(day sharedmodels.Day) bool {
	switch day {
	case sharedmodels.Day(time.Monday):
		return ds.Monday
	case sharedmodels.Day(time.Tuesday):
		return ds.Tuesday
	case sharedmodels.Day(time.Wednesday):
		return ds.Wednesday
	case sharedmodels.Day(time.Thursday):
		return ds.Thursday
	case sharedmodels.Day(time.Friday):
		return ds.Friday
	}
	return false
}

func (v *DaysFormView) ToBitmask() sharedmodels.DaysBitmask {
	mask := 0
	for _, d := range sharedmodels.Days {
		if v.IsDaySet(d) {
			mask |= d.Bit()
		}
	}
	return sharedmodels.DaysBitmask(mask)
}

func DaysBitmaskToFormView(m sharedmodels.DaysBitmask) DaysFormView {
	ds := DaysFormView{}
	for _, d := range sharedmodels.Days {
		ds.setDay(d, m.IsDaySet(d))
	}
	return ds
}

func DaysBitmaskToInitial(m sharedmodels.DaysBitmask) string {
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
