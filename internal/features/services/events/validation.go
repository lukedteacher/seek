package events

import (
	"seek/internal/features/services/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(sm *models.Service) map[string]Validation {
	errors := make(map[string]Validation)
	return errors
}
