package httpserver

import (
	"fmt"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/teachers/blocks"
	"seek/internal/features/teachers/events"
	"seek/internal/features/teachers/models"
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
	user := currentUser(r)
	teachers, err := s.Teachers.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "teachers list db list", "err", err)
		return
	}

	_ = pages.List(user, teachers).Render(ctx, w)
}

// GET request to /teachers/{id}
func (s Server) getTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	teacherID := chi.URLParam(r, "id")
	teacher, err := s.Teachers.Get(ctx, teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.View(user, teacher).Render(ctx, w)
}

// GET request to /teachers/create
func (s Server) getCreateTeacherForm(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	_ = pages.Create(user).Render(r.Context(), w)
}

// POST request to /teachers/create
func (s Server) createTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	type Signals struct {
		GivenName  string `json:"given_name"`
		ChosenName string `json:"chosen_name"`
		FamilyName string `json:"family_name"`
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
	result, err := events.CreateTeacherCommandHandler(ctx, events.CreateTeacherCommand{
		GivenName:  signals.GivenName,
		ChosenName: signals.ChosenName,
		FamilyName: signals.FamilyName,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}

	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/teachers/%s", result.EventID))
}

// GET request to /students/{id}/edit
func (s Server) getEditTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	teacherID := chi.URLParam(r, "id")
	teacherRes, err := s.Teachers.Get(ctx, teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.Logger.ErrorContext(ctx, "get edit teacher db get", "err", err)
		return
	}

	teacherSignals := models.TeacherSignals{
		ID:         teacherRes.ID,
		GivenName:  teacherRes.GivenName,
		ChosenName: *teacherRes.ChosenName,
		FamilyName: teacherRes.FamilyName,
	}

	validation := events.Validate(teacherRes)
	_ = pages.Edit(user, teacherSignals, validation).Render(ctx, w)
}

// POST request to /teachers/{id}/edit/validate
func (s Server) postValidateEditTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teacherID := chi.URLParam(r, "id")
	type Signals struct {
		Teacher models.TeacherSignals `json:"teacher"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post edit teacher validate read signals", "err", err)
		return
	}

	model := models.Teacher{
		ID:         teacherID,
		GivenName:  signals.Teacher.GivenName,
		ChosenName: &signals.Teacher.ChosenName,
		FamilyName: signals.Teacher.FamilyName,
	}

	validation := events.Validate(&model)
	patchTempl(w, r, blocks.EditForm(signals.Teacher, validation))
}

// POST request to /teachers/{id}/edit
func (s Server) postEditTeacher(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	type Signals struct {
		Teacher models.TeacherSignals `json:"teacher"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post edit teacher read signals", "err", err)
		return
	}

	command := events.UpdateTeacherCommand{
		TeacherID:  signals.Teacher.ID,
		GivenName:  signals.Teacher.GivenName,
		ChosenName: signals.Teacher.ChosenName,
		FamilyName: signals.Teacher.FamilyName,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.UpdateTeacherCommandHandler(ctx, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post edit teacher command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.InfoContext(ctx, "post edit teacher", "skipped", result.Skipped)
		return
	}
}

// POST request to /teachers/{id}
func (s Server) deleteTeacher(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	teacherID := chi.URLParam(r, "id")
	_, err := events.DeleteTeacherCommandHandler(r.Context(), events.DeleteTeacherCommand{
		TeacherID: teacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
