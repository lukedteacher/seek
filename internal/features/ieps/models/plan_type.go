package models

import (
	"encoding/json"
	"strconv"
)

type IEPType int

const (
	IEPTypeUnknown IEPType = iota - 1
	IEPTypeInitial
	IEPTypeAnnual
	IEPTypeInterim
)

var IEPTypeList = []IEPType{
	IEPTypeInitial,
	IEPTypeAnnual,
	IEPTypeInterim,
}

func (iepType IEPType) Word() string {
	wordMap := map[IEPType]string{
		-1: "unknown",
		0:  "initial",
		1:  "annual",
		2:  "interim",
	}
	return wordMap[iepType]
}

func (iepType IEPType) String() string {
	strMap := map[IEPType]string{
		-1: "-1",
		0:  "0",
		1:  "1",
		2:  "2",
	}
	return strMap[iepType]
}

func (iepType *IEPType) UnmarshalJSON(b []byte) error {
	// try as number
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*iepType = IEPType(i)
		return nil
	}
	// try as string
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*iepType = IEPType(i)
	return nil
}
