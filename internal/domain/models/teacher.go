package models

import "strings"

type Teacher struct {
	Id          string  `db:"id"`
	FirstName   string  `db:"first_name"`
	ChosenName  *string `db:"chosen_name"`
	LastName    string  `db:"last_name"`
}

func (s Teacher) DisplayName() string {
	if s.ChosenName != nil {
		return *s.ChosenName
	} else {
		return s.FirstName
	}
}

func (s Teacher) FirstL() string {
	return s.DisplayName() + " " + string(s.LastName[0])
}

func (s *Teacher) FirstLast() string {
	return s.DisplayName() + " " + s.LastName
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