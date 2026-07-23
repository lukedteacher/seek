package events

import (
	"seek/internal/features/educators/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(educator *models.Educator) map[string]Validation {
	errors := make(map[string]Validation)
	return errors
}
