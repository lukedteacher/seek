package httpserver

import (
	"context"
	"net/http"
	"strconv"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/students/events"
	"seek/internal/features/students/pages"
	"seek/internal/views/dto"
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
	studentViews := make([]dto.StudentView, len(students))
	for i := range students {
		studentViews[i] = *dto.NewStudentViewFromModel(&students[i])
	}
	_ = pages.List(user, signals.View, studentViews).Render(ctx, w)
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
			studentViews := make([]dto.StudentView, len(students))
			for i := range students {
				studentView := dto.NewStudentViewFromModel(&students[i])
				studentViews[i] = *studentView
			}

			sse.PatchElementTempl(pages.List(user, 0, studentViews))
		}
	}
}

// GET request to /students/create
func (s Server) getStudentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.NewStudent()
	view := dto.NewStudentFormViewFromModel(empty)
	_ = pages.Create(user, *view).Render(ctx, w)
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
			println("watcher update")
			if !ok {
				return
			}
			var model models.Student
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewStudentFormViewFromModel(&model)
			sse.PatchElementTempl(pages.Create(user, *view))
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

	var grade int64
	if signals.Student.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Student.Grade, 10, 64)
	}

	_, err := events.CreateStudentCommandHandler(ctx, events.CreateStudentCommand{
		FirstName:   signals.Student.FirstName,
		ChosenName:  signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
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
	student, err := s.Students.Get(r.Context(), studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewStudentViewFromModel(student)
	periodIDs, _ := s.PeriodsStudents.ListPeriodIDsForStudent(ctx, studentID)
	periodViews := make([]dto.PeriodView, len(periodIDs))
	for i := range periodIDs {
		period, _ := s.Periods.Get(ctx, periodIDs[i])
		view, _ := dto.NewViewFromPeriod(period)
		periodViews[i] = view
	}
	view.Schedule.Periods = periodViews
	_ = pages.View(user, *view).Render(r.Context(), w)
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
			var model models.Student
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewStudentViewFromModel(&model)
			periodIDs, _ := s.PeriodsStudents.ListPeriodIDsForStudent(ctx, studentID)
			periodViews := make([]dto.PeriodView, len(periodIDs))
			for i := range periodIDs {
				period, _ := s.Periods.Get(ctx, periodIDs[i])
				view, _ := dto.NewViewFromPeriod(period)
				periodViews[i] = view
			}
			view.Schedule.Periods = periodViews
			sse.PatchElementTempl(pages.View(user, *view))
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

	view := dto.NewStudentFormViewFromModel(model)
	_ = pages.Edit(user, *view).Render(ctx, w)
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
			var model models.Student
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewStudentFormViewFromModel(&model)
			sse.PatchElementTempl(pages.Edit(user, *view))
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
	view := dto.NewStudentFormViewFromModel(model)
	_ = pages.Edit(user, *view).Render(ctx, w)
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

	gradeInt64, err := strconv.ParseInt(signals.Student.Grade, 10, 64)
	if err != nil {
		println("convert int64 error: ", err.Error())
	}

	studentID := chi.URLParam(r, "id")
	result, err := events.UpdateStudentCommandHandler(ctx, events.UpdateStudentCommand{
		Id:          studentID,
		FirstName:   signals.Student.FirstName,
		ChosenName:  signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
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
	user := currentUser(r)
	studentID := chi.URLParam(r, "id")
	_, err := events.DeleteStudentCommandHandler(r.Context(), events.DeleteStudentCommand{
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
