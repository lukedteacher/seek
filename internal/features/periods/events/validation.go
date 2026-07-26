package events

import (
	"errors"
	"strings"

	"seek/internal/features/periods/models"
	"seek/pkg/utils"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(p *models.Period) map[string]Validation {
	errors := make(map[string]Validation)
	if p == nil {
		return errors
	}
	errors["title"] = ValidateTitle(p.Title)

	if p.StartTime.IsZero() {
		errors["startTime"] = Validation{Message: "required", State: "empty"}
	} else {
		errors["startTime"] = Validation{Message: "required", State: "valid"}
	}

	if p.Duration < 0 {
		errors["duration"] = Validation{Message: "1-60 minutes", State: "error"}
	} else {
		errors["duration"] = Validation{Message: "1-60 minutes", State: "valid"}
	}

	if p.Days > 16 {
		errors["class"] = Validation{Message: "required", State: "empty"}
	} else {
		errors["class"] = Validation{Message: "required", State: "valid"}
	}

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

func ValidateStartTime(startTime string) (string, error) {
	startTime = strings.TrimSpace(startTime)
	if startTime == "" || len(startTime) > 160 {
		return "", errors.New("period start time must be between 1 and 160 characters")
	}
	return startTime, nil
}

func ValidateDuration(duration int64) (int64, error) {
	if duration < 1 || duration > 60 {
		return 0, errors.New("period duration must be between 1 and 60 minutes")
	}
	return duration, nil
}

func ValidateDays(days int64) (int64, error) {
	if days < 0 || days > 16 {
		return 0, errors.New("period days must be between 0 and 16")
	}
	return days, nil
}
