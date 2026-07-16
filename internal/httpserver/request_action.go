package httpserver

import (
	"net/http"

	"seek/internal/eventstore"
)

func setRequestAction(r *http.Request, action string, fields map[string]any) {
	eventstore.SetRequestAction(r.Context(), action, fields)
}
