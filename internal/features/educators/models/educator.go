package models

import (
	"errors"
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

// sentinel errors for Educator model validation
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

func NewEducator() *Educator {
	return &Educator{}
}
