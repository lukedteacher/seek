package httpserver

import (
	"net/http"

	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
)

func (s Server) coreRoutes(r chi.Router) {
	r.Get("/", s.index)
}

func (s Server) index(w http.ResponseWriter, r *http.Request) {
	_ = pages.Index().Render(r.Context(), w)
}