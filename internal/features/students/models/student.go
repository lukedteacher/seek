package models

import (
	"seek/internal/features/_shared/sharedmodels"
)

type Student struct {
	sharedmodels.Person        // embeds given, chosen, & family name fields
	ID                  string `json:"id" display:"ID"`
	Grade               int    `json:"grade" display:"grade" format:"GradeOrdinal" renderer:"badge"`
	Homeroom            string `json:"homeroom" display:"homeroom"`
	CaseManager         string `json:"case_manager" display:"case manager"`
}

func NewStudent() *Student {
	return &Student{}
}

func (s Student) GradeOrdinal() string {
	ordinalMap := map[int]string{
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

	return ordinalMap[s.Grade]
}

func (s Student) GradeWord() string {
	wordMap := map[int]string{
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

	return wordMap[s.Grade]
}

func (s Student) GradeString() string {
	stringMap := map[int]string{
		0:  "0",
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "44",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "10",
		11: "11",
		12: "12",
	}

	return stringMap[s.Grade]
}
