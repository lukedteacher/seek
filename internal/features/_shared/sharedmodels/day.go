package sharedmodels

import (
	"strings"
	"time"
)

type Day time.Weekday

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

func (d Day) Bit() int {
	bits := map[time.Weekday]int{
		time.Monday:    BitMonday,
		time.Tuesday:   BitTuesday,
		time.Wednesday: BitWednesday,
		time.Thursday:  BitThursday,
		time.Friday:    BitFriday,
	}
	return bits[time.Weekday(d)]
}
