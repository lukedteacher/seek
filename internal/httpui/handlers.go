package httpui

import (
	"net/http"

	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
)

func (s Server) routes(r chi.Router) {
	r.Get("/", s.index)
	r.Get("/students", s.students)
}

func (s Server) index(w http.ResponseWriter, r *http.Request) {
	_ = pages.Index().Render(r.Context(), w)
}

func (s Server) students(w http.ResponseWriter, r *http.Request) {
	students, err := s.Students.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.Students(1, students).Render(r.Context(), w)
}