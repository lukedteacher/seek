package models

import (
	"errors"
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

var (
	ErrEmailRequired = errors.New("educator email is required")
	ErrRoleRequired  = errors.New("educator role is required")
)

type Educator struct {
	sharedmodels.Person //embeds given, chosen, family name & email, username fields
	ID                  string
	Roles               []sharedmodels.EducatorRole
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewEducator(id, given, chosen, family, email string, roles []string) (Educator, error) {
	person, err := sharedmodels.NewPerson(given, chosen, family, email)
	if err != nil {
		return Educator{}, err
	}
	return Educator{
		Person: person,
		ID:     id,
	}, nil
}
