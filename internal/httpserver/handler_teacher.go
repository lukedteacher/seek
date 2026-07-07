package httpserver

import (
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/teacher"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) teacherRoutes(r chi.Router) {
	r.Get("/teachers", s.getTeachersList)
	r.Get("/teachers/stream", s.teachersStream)
	r.Get("/teachers/{id}", s.getTeacher)
	r.Get("/teachers/create", s.getCreateTeacherForm)
	r.Post("/teachers/create", s.createTeacher)
	r.Delete("/teachers/{id}", s.deleteTeacher)
}

// GET request to /teachers/{id}
func (s Server) getTeacher(w http.ResponseWriter, r *http.Request) {
	teacherID := chi.URLParam(r, "id")
	teacher, err := s.Teachers.Get(r.Context(), teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.Teacher(teacher).Render(r.Context(), w)
}

// GET request to /teachers
func (s Server) getTeachersList(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		View int64 `json:"view"`
	}
	signals := &Signals{}
	teachers, err := s.Teachers.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.ReadSignals(r, signals)

	_ = pages.Teachers(signals.View, teachers).Render(r.Context(), w)
}

func (s Server) teachersStream(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	ctx := r.Context()

	watcher, err := s.ViewStore.Watch(ctx, "teachers", viewstore.WatchOptions{IgnoreDeletes: true})
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
			teachers, err := s.Teachers.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			page := pages.Teachers(0, teachers)
			if err := sse.PatchElementTempl(page); err != nil {
				return
			}
		}
	}
}

// GET request to /teachers/create
func (s Server) getCreateTeacherForm(w http.ResponseWriter, r *http.Request) {
	_ = pages.CreateTeacher().Render(r.Context(), w)
}

// POST request to /teachers/create
func (s Server) createTeacher(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		FirstName  string `json:"first_name"`
		ChosenName string `json:"chosen_name"`
		LastName   string `json:"last_name"`
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
	_, err := teacher.CreateTeacherCommandHandler(r.Context(), teacher.CreateTeacherCommand{
		FirstName:  signals.FirstName,
		ChosenName: signals.ChosenName,
		LastName:   signals.LastName,
		Metadata:   eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
	
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&Signals{}, sse)
	})
}

// POST request to /teachers/{id}/delete
func (s Server) deleteTeacher(w http.ResponseWriter, r *http.Request) {
	teacherID := chi.URLParam(r, "id")
	_, err := teacher.DeleteTeacherCommandHandler(r.Context(), teacher.DeleteTeacherCommand{
		TeacherID: teacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
