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

func DaysBitmaskToColumnNumbers(mask int64) []int {
	columnNumbers := []int{}
	for i, day := range Days {
		if day.IsSet(mask) {
			columnNumbers = append(columnNumbers, i+1)
		}
	}
	return columnNumbers
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

func (ds *DaysSignals) IsDaySet(day Day) bool {
	switch day {
	case Day(time.Monday):
		return ds.Monday
	case Day(time.Tuesday):
		return ds.Tuesday
	case Day(time.Wednesday):
		return ds.Wednesday
	case Day(time.Thursday):
		return ds.Thursday
	case Day(time.Friday):
		return ds.Friday
	}
	return false
}

func DaysBitmaskToDaysSignals(mask int64) DaysSignals {
	ds := DaysSignals{}
	for _, d := range Days {
		set := mask&d.Bit() != 0
		switch d {
		case Day(time.Monday):
			ds.Monday = set
		case Day(time.Tuesday):
			ds.Tuesday = set
		case Day(time.Wednesday):
			ds.Wednesday = set
		case Day(time.Thursday):
			ds.Thursday = set
		case Day(time.Friday):
			ds.Friday = set
		}
	}
	return ds
}

func DaysSignalsToDaysBitmask(ds DaysSignals) int64 {
	mask := int64(0)
	for _, d := range Days {
		var set bool
		switch d {
		case Day(time.Monday):
			set = ds.Monday
		case Day(time.Tuesday):
			set = ds.Tuesday
		case Day(time.Wednesday):
			set = ds.Wednesday
		case Day(time.Thursday):
			set = ds.Thursday
		case Day(time.Friday):
			set = ds.Friday
		}
		if set {
			mask |= d.Bit()
		}
	}
	return mask
}
