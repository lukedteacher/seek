package httpui

import (
	"net/http"

	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students", s.students)
	r.Get("/students/stream", s.studentsStream)
}

func (s Server) students(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		View int64 `json:"view"`
	}
	signals := &Signals{}
	students, err := s.Students.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.ReadSignals(r, signals)
	
	_ = pages.Students(signals.View, students).Render(r.Context(), w)
}

func (s Server) studentsStream(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	ctx := r.Context()

	watcher, err := s.ViewStore.Watch(ctx, "students", viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		_ = alert(sse, err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watcher.Updates():
			if !ok {
				return
			}
			students, err := s.Students.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			page := pages.Students(0, students)
			if err := sse.PatchElementTempl(page); err != nil {
				return
			}
		}
	}
}