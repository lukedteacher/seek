package httpserver

import (
	"context"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/students/dto"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
	"seek/internal/features/students/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students/list", s.getStudentsList)
	r.Get("/students/list/stream", s.getStudentsListStream)
	r.Get("/students/create", s.getStudentCreate)
	r.Get("/students/create/stream", s.getStudentCreateStream)
	r.Post("/students/create/validate", s.postStudentCreateValidate)
	r.Post("/students/create", s.postStudentCreate)
	r.Get("/students/{id}/view", s.getStudentView)
	r.Get("/students/{id}/view/stream", s.getStudentViewStream)
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
	signals := &struct {
		View int `json:"view"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("signal read error: ", err.Error())
		return
	}
	students, err := s.Students.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewStudentTableView(students)
	_ = pages.List(user, view).Render(ctx, w)
}

// GET request to /students/list/stream
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
		println("students list stream error: ", err.Error())
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
		println("watcher error in student create stream: ", err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			println("student watcher update")
			if !ok {
				return
			}
			var model *models.Student
			if err := entry.JSON(model); err != nil {
				println(err.Error())
				return
			}
			println(model.Grade)
			view := dto.NewStudentFormView(model)
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
		println("pcv signal read: ", err.Error())
		return
	}
	model := dto.NewStudentModelFromView(&signals.Student)
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "newstudent", model); err != nil {
		println("view store error ", err.Error())
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
		println(err.Error())
		return
	}

	var grade int64 = -1

	_, err := events.CreateStudentCommandHandler(ctx, events.CreateStudentCommand{
		GivenName:   signals.Student.GivenName,
		ChosenName:  signals.Student.ChosenName,
		FamilyName:  signals.Student.FamilyName,
		Grade:       grade,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
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
	_ = pages.View(user, view).Render(ctx, w)
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
		println(err.Error())
		return
	}
	defer sub.Close()

	if err := s.refreshStudentViewState(ctx, studentID); err != nil {
		println("svs first refresh: ", err.Error())
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
		println(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshStudentViewState(ctx, studentID); err != nil {
				println("svs second refresh: ", err.Error())
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			var model *models.Student
			if err := entry.JSON(model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewStudentView(model)
			sse.PatchElementTempl(pages.View(user, view))
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
	_ = pages.Edit(user, view).Render(ctx, w)
}

// GET request to /student/{id}/stream
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
		println(err.Error())
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
		println(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshStudentEditState(ctx, studentID); err != nil {
				println(err.Error())
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model *models.Student
			if err := entry.JSON(model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewStudentFormView(model)
			sse.PatchElementTempl(pages.Edit(user, view))
		}
	}
}

// POST request to /students/{id}/edit/validate
func (s Server) postStudentEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := dto.NewStudentModelFromView(&signals.Student)
	model.ID = chi.URLParam(r, "id")
	view := dto.NewStudentFormView(&model)
	_ = pages.Edit(user, view).Render(ctx, w)
}

// POST request to /students/{id}/edit
func (s Server) postStudentEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("error reading signals: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gradeInt64 := int64(signals.Student.Grade)

	studentID := chi.URLParam(r, "id")
	result, err := events.UpdateStudentCommandHandler(ctx, events.UpdateStudentCommand{
		Id:          studentID,
		GivenName:   signals.Student.GivenName,
		ChosenName:  signals.Student.ChosenName,
		FamilyName:  signals.Student.FamilyName,
		Grade:       gradeInt64,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		println("command error: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Skipped == true {
		println("update skipped")
		return
	}
}

// POST request to /students/{id}/delete
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
