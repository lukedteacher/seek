package models

import (
	"encoding/json"
	"strconv"
)

type DisabilityCode int

const (
	DisabilityCodeUnknown DisabilityCode = iota - 1
	DisabilityCodeND
	DisabilityCodeSLI
	DisabilityCodeDCDMild
	DisabilityCodeDCDSevere
	DisabilityCodePI
	DisabilityCodeDHH
	DisabilityCodeVI
	DisabilityCodeSLD
	DisabilityCodeEBD
	DisabilityCodeDB
	DisabilityCodeOHD
	DisabilityCodeASD
	DisabilityCodeDD
	DisabilityCodeTBI
	DisabilityCodeSMI
)

var DisabilityCodeList = []DisabilityCode{
	DisabilityCodeND,
	DisabilityCodeSLI,
	DisabilityCodeDCDMild,
	DisabilityCodeDCDSevere,
	DisabilityCodePI,
	DisabilityCodeDHH,
	DisabilityCodeVI,
	DisabilityCodeSLD,
	DisabilityCodeEBD,
	DisabilityCodeDB,
	DisabilityCodeOHD,
	DisabilityCodeASD,
	DisabilityCodeDD,
	DisabilityCodeTBI,
	DisabilityCodeSMI,
}

func (dc DisabilityCode) Short() string {
	wordMap := map[DisabilityCode]string{
		-1: "unknown",
		0:  "ND",
		1:  "SLI",
		2:  "DCD mild",
		3:  "DCD severe",
		4:  "PI",
		5:  "DHH",
		6:  "VI",
		7:  "SLD",
		8:  "EBD",
		9:  "DB",
		10: "OHD",
		11: "ASD",
		12: "DD",
		13: "TBI",
		14: "SMI",
	}
	return wordMap[dc]
}

func (dc DisabilityCode) Long() string {
	wordMap := map[DisabilityCode]string{
		-1: "unknown",
		0:  "non-disabled",
		1:  "speech-language impaired",
		2:  "developmental / cognitive delay (mild-moderate)",
		3:  "developmental / cognitive delay (severe-profound)",
		4:  "physically impaired",
		5:  "deaf-hard of hearing",
		6:  "visually impaired",
		7:  "specific learning disabilities",
		8:  "emotional / behavior disorders",
		9:  "deaf-blind",
		10: "other health disabilities",
		11: "autism spectrum disorder",
		12: "developmental delay",
		13: "traumatic brain injury",
		14: "severly multiply impaired",
	}
	return wordMap[dc]
}

func (dc DisabilityCode) String() string {
	strMap := map[DisabilityCode]string{
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
		13: "13",
		14: "14",
	}
	return strMap[dc]
}

func (dc *DisabilityCode) UnmarshalJSON(b []byte) error {
	// try as number
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*dc = DisabilityCode(i)
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
	*dc = DisabilityCode(i)
	return nil
}
