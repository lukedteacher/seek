package sharedmodels

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type PlanType int

const (
	PlanTypeNone PlanType = iota
	PlanTypeChildFind
	PlanType504
	PlanTypeIEP
)

var PlanTypeList = []PlanType{
	PlanTypeNone,
	PlanTypeChildFind,
	PlanType504,
	PlanTypeIEP,
}

func (pt PlanType) Int() int {
	return int(pt)
}

func (pt PlanType) String() string {
	stringMap := map[PlanType]string{
		0: "0",
		1: "1",
		2: "2",
		3: "3",
	}
	return stringMap[pt]
}

func (pt PlanType) Description() string {
	stringMap := map[PlanType]string{
		0: "none",
		1: "child find",
		2: "504",
		3: "IEP",
	}
	return stringMap[pt]
}

// MarshalJSON returns the number string.
func (pt PlanType) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// UnmarshalJSON accepts a string and maps it to PlanType.
func (pt *PlanType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*pt = PlanTypeNone
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if v < 0 || v >= len(PlanTypeList) {
		return fmt.Errorf("invalid plan type: %d", v)
	}
	*pt = PlanType(v)
	return nil
}
