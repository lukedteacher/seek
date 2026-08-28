package sharedmodels

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Grade int

const (
	GradeUnassigned Grade = iota - 1
	GradeK
	Grade1
	Grade2
	Grade3
	Grade4
	Grade5
	Grade6
	Grade7
	Grade8
	Grade9
	Grade10
	Grade11
	Grade12
)

var GradeList = []Grade{
	GradeK,
	Grade1,
	Grade2,
	Grade3,
	Grade4,
	Grade5,
	Grade6,
	Grade7,
	Grade8,
}

func (g Grade) Ordinal() string {
	ordinalMap := map[Grade]string{
		-1: "unassigned",
		0:  "K",
		1:  "1st",
		2:  "2nd",
		3:  "3rd",
		4:  "4th",
		5:  "5th",
		6:  "6th",
		7:  "7th",
		8:  "8th",
		9:  "9th",
		10: "10th",
		11: "11th",
		12: "12th",
	}
	return ordinalMap[g]
}

func (g Grade) Word() string {
	wordMap := map[Grade]string{
		-1: "unassigned",
		0:  "kindergarten",
		1:  "first",
		2:  "second",
		3:  "third",
		4:  "fourth",
		5:  "fifth",
		6:  "eighth",
		7:  "seventh",
		8:  "eighth",
		9:  "ninth",
		10: "tenth",
		11: "eleventh",
		12: "twelfth",
	}
	return wordMap[g]
}

func (g Grade) String() string {
	strMap := map[Grade]string{
		-1: "-1",
		0:  "0",
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "4",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "10",
		11: "11",
		12: "12",
	}
	return strMap[g]
}

func (g *Grade) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*g = Grade(int(val))
		return nil
	case string:
		if val == "" {
			*g = Grade(-1)
			return nil
		}
		i, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*g = Grade(i)
		return nil
	default:
		return fmt.Errorf("grade: unsupported type %T", v)
	}
}
