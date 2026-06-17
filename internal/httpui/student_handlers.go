package httpui

import (
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/student"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students", s.students)
	r.Get("/students/stream", s.studentsStream)
	r.Post("/student", s.createStudent)
	r.Get("/student/create", s.createStudentForm)
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

// get request to /student/create
func (s Server) createStudentForm(w http.ResponseWriter, r *http.Request) {
	_ = pages.CreateStudent().Render(r.Context(), w)
}

// post request to /student
func (s Server) createStudent(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    FirstName    string `json:"first_name"`
    ChosenName   string `json:"chosen_name"`
    LastName     string `json:"last_name"`
    Grade        int64	`json:"grade"`
    Homeroom     string `json:"homeroom"`
    CaseManager  string `json:"case_manager"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }

	println(signals.Grade)
  
  _, err := student.CreateStudentCommandHandler(r.Context(), student.CreateStudentCommand{
    FirstName:    signals.FirstName,
    ChosenName:   signals.ChosenName,
    LastName:     signals.LastName,
    Grade:        signals.Grade,
    Homeroom:     signals.Homeroom,
    CaseManager:  signals.CaseManager,
    Metadata:     eventstore.HTTPCommandMetadata(r),
  }, s.EventSaver)
  if err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
  
  writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
    return clearNewStudentForm(sse)
  })
}