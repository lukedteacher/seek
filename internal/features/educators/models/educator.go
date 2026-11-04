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

func (e Educator) IsFresh() bool {
	return time.Since(e.UpdatedAt) < 2*time.Second
}

func (e Educator) IsCaseManager() bool {
	for _, role := range e.Roles {
		if role == sharedmodels.EducatorRoleCaseManager {
			return true
		}
	}
	return false
}
