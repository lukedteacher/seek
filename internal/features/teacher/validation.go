package teacher

import (
	"seek/internal/domain/models"
	"seek/utils"
)

type Validation struct {
	Message string
	State   string // "error", "empty", "valid"
}

func Validate(teacher *models.Teacher) map[string]Validation {
	errors := make(map[string]Validation)
	if teacher == nil {
		println("am i sending a nil teacher?")
	}
	if teacher.ChosenName == nil {
		println("why does this fial?")
	}
	println("test")
	errors["first_name"] = ValidateFirstName(teacher.FirstName)
	println("test2")
	errors["chosen_name"] = ValidateChosenName(*teacher.ChosenName)
	println("test3")
	errors["last_name"] = ValidateLastName(teacher.LastName)
	println("test4")

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
	println("am i getting here??")
	if chosenName == "" {
		return Validation{Message: "optional", State: "empty"}
	} else	if utils.ValidateAlphanumericLax(chosenName) {
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
