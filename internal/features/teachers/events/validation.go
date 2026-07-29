package events

import (
	"seek/internal/features/teachers/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(teacher *models.Teacher) map[string]Validation {
	errors := make(map[string]Validation)
	return errors
}
