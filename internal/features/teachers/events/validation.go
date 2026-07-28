package events

import (
	"seek/internal/features/teachers/models"
	"seek/pkg/utils"
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
	errors["given_name"] = ValidateGivenName(teacher.GivenName)
	println("test2")
	errors["chosen_name"] = ValidateChosenName(*teacher.ChosenName)
	println("test3")
	errors["family_name"] = ValidateFamilyName(teacher.FamilyName)
	println("test4")

	return errors
}

func ValidateGivenName(givenName string) Validation {
	if givenName == "" {
		return Validation{Message: "required", State: "empty"}
	} else if utils.ValidateAlphanumeric(givenName) {
		return Validation{Message: "required", State: "valid"}
	} else {
		return Validation{Message: "required", State: "error"}
	}
}

func ValidateChosenName(chosenName string) Validation {
	println("am i getting here??")
	if chosenName == "" {
		return Validation{Message: "optional", State: "empty"}
	} else if utils.ValidateAlphanumericLax(chosenName) {
		return Validation{Message: "optional", State: "valid"}
	} else {
		return Validation{Message: "optional", State: "error"}
	}
}

func ValidateFamilyName(familyName string) Validation {
	if familyName == "" {
		return Validation{Message: "required", State: "empty"}
	} else if utils.ValidateAlphanumeric(familyName) {
		return Validation{Message: "required", State: "valid"}
	} else {
		return Validation{Message: "required", State: "error"}
	}
}
