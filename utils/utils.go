package utils

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func ChiParamInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func ValidateAlphanumeric(i string) bool {
	var rxPat = regexp.MustCompile(`^[0-9A-Za-z ]+$`)
	return rxPat.MatchString(i)
}

func ValidateAlphanumericLax(i string) bool {
	var rxPat = regexp.MustCompile(`^[0-9A-Za-z !.?]+$`)
	return rxPat.MatchString(i)
}