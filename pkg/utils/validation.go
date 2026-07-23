package utils

import (
	"regexp"
)

func ValidateAlphanumeric(i string) bool {
	var rxPat = regexp.MustCompile(`^[0-9A-Za-z ]+$`)
	return rxPat.MatchString(i)
}

func ValidateAlphanumericLax(i string) bool {
	var rxPat = regexp.MustCompile(`^[0-9A-Za-z !.?]+$`)
	return rxPat.MatchString(i)
}
