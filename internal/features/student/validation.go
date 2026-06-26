package student

import (
	"seek/internal/domain/models"
	"seek/utils"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(student *models.Student) map[string]Validation {
	errors := make(map[string]Validation)
	if student == nil {
		return errors
	}
	errors["first_name"] = ValidateFirstName(student.FirstName)
	errors["chosen_name"] = ValidateChosenName(*student.ChosenName)
	errors["last_name"] = ValidateLastName(student.LastName)
	errors["grade"] = ValidateGrade(student.Grade)

	return errors
}

func ValidateFirstName(firstName string) Validation {
	if firstName == "" {
		return Validation{Message: "required", State: "empty"}
	} else if utils.ValidateAlphanumeric(firstName) {
		return Validation{Message: "required", State: "valid"}
	} else {
		return Validation{Message: "required", State: "error"}
	}
}

func ValidateChosenName(chosenName string) Validation {
	if chosenName == "" {
		return Validation{Message: "optional", State: "empty"}
	} else	if utils.ValidateAlphanumeric(chosenName) {
		return Validation{Message: "optional", State: "valid"}
	} else {
		return Validation{Message: "optional", State: "error"}
	}
}

func ValidateLastName(lastName string) Validation {
	if lastName == "" {
		return Validation{Message: "required", State: "empty"}
	} else if utils.ValidateAlphanumeric(lastName) {
		return Validation{Message: "required", State: "valid"}
	} else {
		return Validation{Message: "required", State: "error"}
	}
}

func ValidateGrade(grade int64) Validation {
	if grade == -1 {
		return Validation{Message: "required", State: "empty"}
	} else {
		return Validation{Message: "required", State: "valid"}
	}
}
