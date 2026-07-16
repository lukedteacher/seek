package httpserver

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/teachers/blocks"
	"seek/internal/features/teachers/events"
	"seek/internal/features/teachers/pages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) teacherRoutes(r chi.Router) {
	r.Get("/teachers", s.getTeachersList)
	r.Get("/teachers/{id}", s.getTeacher)
	r.Get("/teachers/create", s.getCreateTeacherForm)
	r.Post("/teachers/create", s.createTeacher)
	r.Get("/teachers/{id}/edit", s.getEditTeacher)
	r.Post("/teachers/{id}/edit/validate", s.postValidateEditTeacher)
	r.Post("/teachers/{id}/edit", s.postEditTeacher)
	r.Delete("/teachers/{id}", s.deleteTeacher)
}

// GET request to /teachers
func (s Server) getTeachersList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Signals struct {
		View int64 `json:"view"`
	}
	signals := &Signals{}
	teachers, err := s.Teachers.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.ReadSignals(r, signals)

	_ = pages.List(signals.View, teachers).Render(ctx, w)
}

// GET request to /teachers/{id}
func (s Server) getTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teacherID := chi.URLParam(r, "id")
	teacher, err := s.Teachers.Get(ctx, teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.View(teacher).Render(ctx, w)
}

// GET request to /teachers/create
func (s Server) getCreateTeacherForm(w http.ResponseWriter, r *http.Request) {
	_ = pages.Create().Render(r.Context(), w)
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
	_, err := events.CreateTeacherCommandHandler(r.Context(), events.CreateTeacherCommand{
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

// GET request to /students/{id}/edit
func (s Server) getEditTeacher(w http.ResponseWriter, r *http.Request) {
	context := r.Context()

	teacherID := chi.URLParam(r, "id")
	teacherRes, err := s.Teachers.Get(context, teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	teacherSignals := models.TeacherSignals{
		ID:         teacherRes.ID,
		FirstName:  teacherRes.FirstName,
		ChosenName: *teacherRes.ChosenName,
		LastName:   teacherRes.LastName,
	}

	validation := events.Validate(teacherRes)
	_ = pages.Edit(teacherSignals, validation).Render(context, w)
}

// POST request to /teachers/{id}/edit/validate
func (s Server) postValidateEditTeacher(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Teacher models.TeacherSignals `json:"teacher"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vet signals: ", err.Error())
		return
	}

	teacherID := chi.URLParam(r, "id")

	model := models.Teacher{
		ID:         teacherID,
		FirstName:  signals.Teacher.FirstName,
		ChosenName: &signals.Teacher.ChosenName,
		LastName:   signals.Teacher.LastName,
	}

	validation := events.Validate(&model)
	patchTempl(w, r, blocks.EditForm(signals.Teacher, validation))
}

// POST request to /teachers/{id}/edit
func (s Server) postEditTeacher(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Teacher models.TeacherSignals `json:"teacher"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := models.Teacher{
		ID:         signals.Teacher.ID,
		FirstName:  signals.Teacher.FirstName,
		ChosenName: &signals.Teacher.ChosenName,
		LastName:   signals.Teacher.LastName,
	}

	// TODO add actual validation
	validation := events.Validate(&model)
	if validation == nil {
		println("some validation error")
	}

	command := events.UpdateTeacherCommand{
		TeacherID:  signals.Teacher.ID,
		FirstName:  signals.Teacher.FirstName,
		ChosenName: signals.Teacher.ChosenName,
		LastName:   signals.Teacher.LastName,
		Metadata:   eventstore.HTTPCommandMetadata(r),
	}
	result, err := events.UpdateTeacherCommandHandler(context, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Skipped == true {
		println("update skipped")
		return
	}
}

// POST request to /teachers/{id}/delete
func (s Server) deleteTeacher(w http.ResponseWriter, r *http.Request) {
	teacherID := chi.URLParam(r, "id")
	_, err := events.DeleteTeacherCommandHandler(r.Context(), events.DeleteTeacherCommand{
		TeacherID: teacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
