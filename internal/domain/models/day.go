package models

import (
	"strings"
	"time"
)

type Day time.Weekday

const (
	BitMonday    = 1 << iota // 1
	BitTuesday               // 2
	BitWednesday             // 4
	BitThursday              // 8
	BitFriday                // 16
)

var Days = []Day{
	Day(time.Monday),
	Day(time.Tuesday),
	Day(time.Wednesday),
	Day(time.Thursday),
	Day(time.Friday),
}

func (d Day) Name() string {
	return time.Weekday(d).String()
}

func (d Day) NameLower() string {
	return strings.ToLower(time.Weekday(d).String())
}

func (d Day) Short() string {
	return time.Weekday(d).String()[:3]
}

func (d Day) Initial() string {
	return string(time.Weekday(d).String()[0])
}

func (d Day) Bit() int64 {
	bits := map[time.Weekday]int64{
		time.Monday:    BitMonday,
		time.Tuesday:   BitTuesday,
		time.Wednesday: BitWednesday,
		time.Thursday:  BitThursday,
		time.Friday:    BitFriday,
	}
	return bits[time.Weekday(d)]
}

func DaysBitmaskToDaysSlice(mask int64) []Day {
	var days []Day
	for _, d := range Days {
		if mask&d.Bit() != 0 {
			days = append(days, d)
		}
	}
	return days
}

func DaysSliceToDaysBitmask(days []Day) int64 {
	var mask int64 = 0
	for _, d := range days {
		mask |= d.Bit()
	}
	return mask
}

func IsDaySet(mask int64, day Day) bool {
	return mask&day.Bit() != 0
}

func (d Day) IsSet(mask int64) bool {
	return mask&d.Bit() != 0
}

type DaysSignals struct {
	Monday    bool `json:"monday"`
	Tuesday   bool `json:"tuesday"`
	Wednesday bool `json:"wednesday"`
	Thursday  bool `json:"thursday"`
	Friday    bool `json:"friday"`
}

func DaysSignalToDaysSlice(ds DaysSignals) []Day {
	days := []Day{}
	if ds.Monday {
		days = append(days, Day(time.Monday))
	}
	if ds.Tuesday {
		days = append(days, Day(time.Tuesday))
	}
	if ds.Wednesday {
		days = append(days, Day(time.Wednesday))
	}
	if ds.Thursday {
		days = append(days, Day(time.Thursday))
	}
	if ds.Friday {
		days = append(days, Day(time.Friday))
	}
	return days
}

func DaysSliceToDaysSignal(days []Day) DaysSignals {
	ds := DaysSignals{}
	for _, day := range days {
		switch day {
		case Days[0]: ds.Monday = true
		case Days[1]: ds.Tuesday = true
		case Days[2]: ds.Wednesday = true
		case Days[3]: ds.Thursday = true
		case Days[4]: ds.Friday = true
		}
	}
	return ds
}

func (ds *DaysSignals) IsDaySet(day Day) bool {
	switch day {
	case Day(time.Monday): return ds.Monday
	case Day(time.Tuesday): return ds.Tuesday
	case Day(time.Wednesday): return ds.Wednesday
	case Day(time.Thursday): return ds.Thursday
	case Day(time.Friday): return ds.Friday
	}
	return false
}

func DaysBitmaskToDaysSignals(mask int64) DaysSignals {
	days := DaysBitmaskToDaysSlice(mask)
	daysSignals := DaysSliceToDaysSignal(days)
	return daysSignals
}

func DaysSignalsToDaysBitmask(ds DaysSignals) int64 {
	days := DaysSignalToDaysSlice(ds)
	mask := DaysSliceToDaysBitmask(days)
	return mask
}