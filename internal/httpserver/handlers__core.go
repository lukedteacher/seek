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
	ctx := r.Context()
	user, _, err := s.Sessions.CurrentUser(ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.Index(user).Render(r.Context(), w)
}

func (s Server) components(w http.ResponseWriter, r *http.Request) {
	_ = pages.Components().Render(r.Context(), w)
}