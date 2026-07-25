package sharedmodels

import (
	"errors"
	"strings"
)

var (
	ErrGivenNameRequired  = errors.New("given name is required")
	ErrFamilyNameRequired = errors.New("family name is required")
	ErrEmailInvalid       = errors.New("invalid email address")
)

type Person struct {
	GivenName  string `json:"given_name" display:"given"`
	ChosenName string `json:"chosen_name" display:"chosen"`
	FamilyName string `json:"family_name" display:"family"`
	Email      string `json:"email" display:"email"`
}

func NewPerson(given, chosen, family, email string) (Person, error) {
	if strings.TrimSpace(given) == "" {
		return Person{}, ErrGivenNameRequired
	}
	if strings.TrimSpace(family) == "" {
		return Person{}, ErrFamilyNameRequired
	}
	if email != "" && !strings.Contains(email, "@") {
		return Person{}, ErrEmailInvalid
	}

	return Person{
		GivenName:  given,
		ChosenName: chosen,
		FamilyName: family,
		Email:      email,
	}, nil
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

// returns the given + family name of a person
func (p Person) LegalFullName() string {
	return p.GivenName + " " + p.FamilyName
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
