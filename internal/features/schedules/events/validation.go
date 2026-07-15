package events

import (
	"seek/internal/domain/models"
	"seek/utils"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(schedule models.Schedule) map[string]Validation {
	errors := make(map[string]Validation)
	errors[ScheduleTitleField] = ValidateTitle(schedule.Title)
	errors[ScheduleTeacherIDField] = ValidateTeacherID(schedule.TeacherId)
	return errors
}

func ValidateTitle(title string) Validation {
	if title == "" {
		return Validation{Message: "required", State: "empty"}
	} else if utils.ValidateAlphanumeric(title) {
		return Validation{Message: "required", State: "valid"}
	} else {
		return Validation{Message: "required", State: "error"}
	}
}

func ValidateTeacherID(teacherID string) Validation {
	if teacherID == "" {
		return Validation{Message: "required", State: "empty"}
	} else	if teacherID == "select a teacher" {
		return Validation{Message: "required", State: "empty"}
	} else {
		return Validation{Message: "required", State: "valid"}
	}
}
