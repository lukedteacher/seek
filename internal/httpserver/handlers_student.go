package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/eventstore"
	idto "seek/internal/features/iepservices/dto"
	"seek/internal/features/students/dto"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
	"seek/internal/features/students/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students", s.getStudentsList)
	r.Get("/students/stream", s.getStudentsListStream)
	r.Get("/students/create", s.getStudentCreate)
	r.Get("/students/create/stream", s.getStudentCreateStream)
	r.Post("/students/create/validate", s.postStudentCreateValidate)
	r.Post("/students/create", s.postStudentCreate)
	r.Get("/students/{id}", s.getStudentView)
	r.Get("/students/{id}/stream", s.getStudentViewStream)
	r.Get("/students/{id}/edit", s.getStudentEdit)
	r.Get("/students/{id}/edit/stream", s.getStudentEditStream)
	r.Post("/students/{id}/edit/validate", s.postStudentEditValidate)
	r.Post("/students/{id}/edit", s.postStudentEdit)
	r.Delete("/students/{id}", s.deleteStudent)
}

// GET request to /students
func (s Server) getStudentsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	students, err := s.Students.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "students list db list", "err", err)
		return
	}
	view := dto.NewStudentTableView(students)
	view.URL = "/students"
	_ = pages.List(user, view).Render(ctx, w)
}

// GET request to /students/stream
func (s Server) getStudentsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)
	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any students
	sub, err := s.Subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "students list stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			// for now just reloads the page
			// consider adding a view store for the list
			students, err := s.Students.List(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			view := dto.NewStudentTableView(students)
			view.URL = "/students"
			sse.PatchElementTempl(pages.List(user, view))
		}
	}
}

// GET request to /students/create
func (s Server) getStudentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.NewStudent()
	view := dto.NewStudentFormView(empty)
	view.URL = "/students/create"
	_ = pages.Create(user, view).Render(ctx, w)
}

// GET request to /students/create/stream
func (s Server) getStudentCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		"newstudent",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student create stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.Student{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "student create stream json", "err", err)
				return
			}
			view := dto.NewStudentFormView(model)
			view.URL = "/students/create"
			sse.PatchElementTempl(pages.Create(user, view))
		}
	}
}

// POST request to /students/create/validate
func (s Server) postStudentCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "student create validate signal read", "err", err)
		return
	}
	model := dto.NewStudentModelFromView(&signals.Student)
	if err := viewstore.PutState(ctx, s.ViewStore, "newstudent", model); err != nil {
		s.Logger.ErrorContext(ctx, "student create validate put state", "err", err)
		return
	}
}

// POST request to /student/create
func (s Server) postStudentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "student create signal read", "err", err)
		return
	}

	result, err := events.CreateStudentCommandHandler(ctx, events.CreateStudentCommand{
		GivenName:   signals.Student.GivenName,
		ChosenName:  signals.Student.ChosenName,
		FamilyName:  signals.Student.FamilyName,
		Grade:       int(signals.Student.Grade),
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student create create command handler", "err", err)
		return
	}

	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/students/%s", result.EventID))
}

// GET request to /students/{id}
func (s Server) getStudentView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	student, err := s.Students.Get(ctx, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewStudentView(student)
	view.URL = fmt.Sprintf("/students/%s", studentID)
	services, err := s.IEPServices.ListIEPServicesForStudent(ctx, studentID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view db list services", "err", err)
	}
	serviceViews := make([]idto.IEPServiceView, len(services))
	for i, service := range services {
		serviceViews[i] = idto.NewIEPServiceView(&service)
	}
	_ = pages.View(user, view, serviceViews).Render(ctx, w)
}

// GET request to /students/{id}/stream
func (s Server) getStudentViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(studentID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshStudentViewState(ctx, studentID); err != nil {
		s.Logger.ErrorContext(ctx, "student view stream refresh", "err", err)
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		studentID+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshStudentViewState(ctx, studentID); err != nil {
				s.Logger.ErrorContext(ctx, "student view stream refresh in select", "err", err)
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.Student{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "student view stream json read", "err", err)
				return
			}
			view := dto.NewStudentView(model)
			view.URL = fmt.Sprintf("/students/%s", studentID)
			services, err := s.IEPServices.ListIEPServicesForStudent(ctx, studentID)
			if err != nil {
				s.Logger.ErrorContext(ctx, "get student view db list services", "err", err)
			}
			serviceViews := make([]idto.IEPServiceView, len(services))
			for i, service := range services {
				serviceViews[i] = idto.NewIEPServiceView(&service)
			}
			sse.PatchElementTempl(pages.View(user, view, serviceViews))
		}
	}
}

// GET request to /students/{id}/edit
func (s Server) getStudentEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	model, err := s.Students.Get(ctx, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}

	view := dto.NewStudentFormView(model)
	view.URL = fmt.Sprintf("/students/%s/edit", studentID)
	_ = pages.Edit(user, view).Render(ctx, w)
}

// GET request to /student/{id}/edit/stream
func (s Server) getStudentEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(studentID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "student edit stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	// watches the student edit view state kv
	watcher, err := s.ViewStore.Watch(
		ctx,
		studentID+".edit",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student edit stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshStudentEditState(ctx, studentID); err != nil {
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				s.Logger.ErrorContext(ctx, "student edit stream refresh in select", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			model := &models.Student{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "student edit stream json read", "err", err)
				return
			}
			view := dto.NewStudentFormView(model)
			view.URL = fmt.Sprintf("/students/%s/edit", studentID)
			sse.PatchElementTempl(pages.Edit(user, view))
		}
	}
}

// POST request to /students/{id}/edit/validate
func (s Server) postStudentEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentID := chi.URLParam(r, "id")
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "student edit validate read signals", "err", err)
		return
	}
	model := dto.NewStudentModelFromView(&signals.Student)
	model.ID = studentID
	viewstore.PutState(ctx, s.ViewStore, studentID, model)
}

// POST request to /students/{id}/edit
func (s Server) postStudentEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post student edit read signals", "err", err)
		return
	}
	result, err := events.UpdateStudentCommandHandler(ctx, events.UpdateStudentCommand{
		StudentID:   studentID,
		GivenName:   signals.Student.GivenName,
		ChosenName:  signals.Student.ChosenName,
		FamilyName:  signals.Student.FamilyName,
		Grade:       int(signals.Student.Grade),
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post student edit command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.InfoContext(ctx, "post student edit command handler", "skipped", result.Skipped)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/students/%s", studentID))
}

// POST request to /students/{id}
func (s Server) deleteStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	_, err := events.DeleteStudentCommandHandler(ctx, events.DeleteStudentCommand{
		StudentID: studentID,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}

// HELPER FUNCTIONS
func (s Server) refreshStudentViewState(ctx context.Context, studentID string) error {
	student, err := s.Students.Get(ctx, studentID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, student.ID+".view", student)
}

func (s Server) refreshStudentEditState(ctx context.Context, studentID string) error {
	student, err := s.Students.Get(ctx, studentID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, student.ID+".edit", student)
}
