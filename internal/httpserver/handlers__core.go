package httpserver

import (
	"net/http"

	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
)

func (s Server) coreRoutes(r chi.Router) {
	r.Get("/", s.index)
	r.Get("/components", s.components)
}

func (s Server) index(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	_ = pages.Index(user).Render(r.Context(), w)
}

func (s Server) components(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	_ = pages.Components(user).Render(r.Context(), w)
}