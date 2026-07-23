package utils

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func ChiParamInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}
