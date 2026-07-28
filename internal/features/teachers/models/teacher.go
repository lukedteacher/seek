package models

import "strings"

type Teacher struct {
	ID         string  `db:"id"`
	GivenName  string  `db:"given_name"`
	ChosenName *string `db:"chosen_name"`
	FamilyName string  `db:"family_name"`
}

type TeacherSignals struct {
	ID         string `json:"id"`
	GivenName  string `json:"given_name"`
	ChosenName string `json:"chosen_name"`
	FamilyName string `json:"family_name"`
}

func (s Teacher) DisplayName() string {
	if s.ChosenName != nil {
		return *s.ChosenName
	} else {
		return s.GivenName
	}
}

func (s Teacher) FirstL() string {
	return s.DisplayName() + " " + string(s.FamilyName[0])
}

func (s *Teacher) FirstLast() string {
	return s.DisplayName() + " " + s.FamilyName
}

func (s *Teacher) Initials() string {
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
