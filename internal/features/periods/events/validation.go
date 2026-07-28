package events

import (
	"seek/internal/features/periods/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(p *models.Period) map[string]Validation {
	errors := make(map[string]Validation)

	return errors
}
