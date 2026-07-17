package events

import (
	"seek/internal/features/iep_services/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(sm *models.IEPService) map[string]Validation {
	errors := make(map[string]Validation)
	return errors
}
