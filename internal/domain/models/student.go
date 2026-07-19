package models

import "strings"

type Student struct {
	ID          string  `json:"id" display:"ID"`
	FirstName   string  `json:"first_name" display:"given"`
	ChosenName  *string `json:"chosen_name" display:"chosen"`
	LastName    string  `json:"last_name" display:"family"`
	Grade       int64   `json:"grade" display:"grade" format:"GradeOrdinal" renderer:"badge"`
	Homeroom    string  `json:"homeroom" display:"homeroom"`
	CaseManager *string `json:"case_manager" display:"case manager"`
}

func NewStudent() *Student {
	return &Student{}
}

func (s Student) DisplayName() string {
	if s.ChosenName != nil && *s.ChosenName != "" {
		return *s.ChosenName
	} else {
		return s.FirstName
	}
}

func (s Student) FirstL() string {
	return s.DisplayName() + " " + string(s.LastName[0])
}

func (s Student) FirstLast() string {
	return s.DisplayName() + " " + s.LastName
}

func (s Student) Initials() string {
	name := s.FirstLast()

	// split by spaces and hyphens
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-'
	})

	// take first character of each part
	var initials []rune
	for _, part := range parts {
		if len(part) > 0 {
			initials = append(initials, []rune(part)[0])
		}
	}

	return string(initials)
}

func (s Student) GradeOrdinal() string {
	ordinalMap := map[int64]string{
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
	wordMap := map[int64]string{
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
	stringMap := map[int64]string{
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
