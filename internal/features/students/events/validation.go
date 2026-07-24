package events

import (
	"seek/internal/features/students/models"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(student *models.Student) map[string]Validation {
	errors := make(map[string]Validation)
	return errors
}
