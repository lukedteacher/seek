package sharedmodels

import "strings"

type Person struct {
	GivenName  string `json:"given_name" display:"given"`
	ChosenName string `json:"chosen_name" display:"chosen"`
	FamilyName string `json:"family_name" display:"family"`
}

// returns the preferred name of a person
func (p Person) Name() string {
	if p.ChosenName != "" {
		return p.ChosenName
	}
	return p.GivenName
}

// returns the preferred name + last initial of a person
func (p Person) NameInitial() string {
	return p.Name() + " " + string([]rune(p.FamilyName)[0])
}

// returns the preferred + family name of a person
func (p Person) FullName() string {
	return p.Name() + " " + p.FamilyName
}

// returns the initials of the preferred + family name of a person
func (p Person) Initials() string {
	name := p.FullName()
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-'
	})

	initials := make([]rune, 0, len(parts))
	for i := range parts {
		initials = append(initials, []rune(parts[i])[0])
	}
	return string(initials)
}
