package sharedmodels

import (
	"fmt"
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

var DayList = []Day{
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

var weekdayInitial = map[time.Weekday]string{
	time.Monday:    "M",
	time.Tuesday:   "T",
	time.Wednesday: "W",
	time.Thursday:  "H",
	time.Friday:    "F",
}

func (d Day) Initial() string {
	return weekdayInitial[time.Weekday(d)]
}

func (d Day) Column() int {
	columns := map[time.Weekday]int{
		time.Monday:    1,
		time.Tuesday:   2,
		time.Wednesday: 3,
		time.Thursday:  4,
		time.Friday:    5,
	}
	return columns[time.Weekday(d)]
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

func ParseDay(s string) (Day, error) {
	switch s {
	case "monday":
		return Day(time.Monday), nil
	case "tuesday":
		return Day(time.Tuesday), nil
	case "wednesday":
		return Day(time.Wednesday), nil
	case "thursday":
		return Day(time.Thursday), nil
	case "friday":
		return Day(time.Friday), nil
	default:
		return 0, fmt.Errorf("invalid day: %q", s)
	}
}
