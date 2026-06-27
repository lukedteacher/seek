package models

import (
	"strings"
	"time"
)

type Day struct {
	Weekday time.Weekday
}

const (
	BitMonday    = 1 << iota // 1
	BitTuesday               // 2
	BitWednesday             // 4
	BitThursday              // 8
	BitFriday                // 16
)

var Days = []Day{
	{time.Monday},
	{time.Tuesday},
	{time.Wednesday},
	{time.Thursday},
	{time.Friday},
}

func (d Day) Name() string {
	return d.Weekday.String()
}
func (d Day) NameLower() string {
	return strings.ToLower(d.Weekday.String())
}

func (d Day) Short() string {
	return d.Weekday.String()[:3]
}

func (d Day) Initial() string {
	return string(d.Weekday.String()[0])
}

func (d Day) Bit() int {
	bits := map[time.Weekday]int{
		time.Monday:    BitMonday,
		time.Tuesday:   BitTuesday,
		time.Wednesday: BitWednesday,
		time.Thursday:  BitThursday,
		time.Friday:    BitFriday,
	}
	return bits[d.Weekday]
}

func DaysFromBitmask(mask int) []Day {
	var days []Day
	for _, d := range Days {
		if mask&d.Bit() != 0 {
			days = append(days, d)
		}
	}
	return days
}

func BitmaskFromDays(days []Day) int {
	mask := 0
	for _, d := range days {
		mask |= d.Bit()
	}
	return mask
}

func IsDaySet(mask int, day Day) bool {
	return mask&day.Bit() != 0
}

func (d Day) IsSet(mask int) bool {
	return mask&d.Bit() != 0
}
