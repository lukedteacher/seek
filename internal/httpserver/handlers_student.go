package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	csevents "seek/internal/features/caseload_students/events"
	edto "seek/internal/features/educators/dto"
	eevents "seek/internal/features/educators/events"
	idto "seek/internal/features/iepservices/dto"
	scheduledto "seek/internal/features/schedules/dto"
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
	r.Get("/students/{username}", s.getStudentView)
	r.Get("/students/{username}/info", s.getStudentViewInfo)
	r.Get("/students/{username}/info/stream", s.getStudentViewInfoStream)
	r.Get("/students/{username}/schedule", s.getStudentViewSchedule)
	r.Get("/students/{username}/schedule/stream", s.getStudentViewScheduleStream)
	r.Get("/students/{username}/services", s.getStudentViewServices)
	r.Get("/students/{username}/services/stream", s.getStudentViewServicesStream)
	r.Get("/students/{username}/edit", getStudentEdit(s.Logger, s.ReadModels.Students))
	r.Get("/students/{username}/edit/stream", getStudentEditStream(s.Logger, s.ViewStore, s.Subscriber, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Post("/students/{username}/edit/validate", s.postStudentEditValidate)
	r.Post("/students/{username}/edit", s.postStudentEdit)
	r.Post("/students/{username}/archive", s.postStudentArchive)
	r.Delete("/students/{username}", s.deleteStudent)
}

// GET request to /students
func (s Server) getStudentsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	students, err := s.ReadModels.Students.List(ctx, events.WithServices())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	studentTableView := dto.NewStudentTableView(students)
	studentTableView.URL = "/students"
	_ = pages.List(studentTableView).Render(ctx, w)
}

// GET request to /students/stream
func (s Server) getStudentsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
			students, err := s.ReadModels.Students.List(ctx, events.WithServices())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			studentTableView := dto.NewStudentTableView(students)
			studentTableView.URL = "/students"
			sse.PatchElementTempl(pages.List(studentTableView))
		}
	}
}

// GET request to /students/create
func (s Server) getStudentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	empty := models.NewStudent()
	studentFormView := dto.NewStudentFormView(empty)
	_ = pages.Create(studentFormView).Render(ctx, w)
}

// GET request to /students/create/stream
func (s Server) getStudentCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
			student := &models.Student{}
			if err := entry.JSON(student); err != nil {
				s.Logger.ErrorContext(ctx, "student create stream json", "err", err)
				return
			}
			studentFormView := dto.NewStudentFormView(student)
			sse.PatchElementTempl(pages.Create(studentFormView))
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
	student := dto.NewStudentModelFromView(&signals.Student)
	if err := viewstore.PutState(ctx, s.ViewStore, "newstudent", student); err != nil {
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
		Email:       signals.Student.Email,
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

// GET request to /students/{username}
// redirects to /students/{username}/info
func (s Server) getStudentView(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	http.Redirect(w, r, fmt.Sprintf("/students/%s/info", username), http.StatusFound)
}

// GET request to /students/{username}/info
func (s Server) getStudentViewInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get the username from the URL and pull the data from the db
	username := chi.URLParam(r, "username")
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view db get", "err", err)
		return
	}

	// create the student view and set the URL
	studentView := dto.NewStudentView(student)

	_ = pages.View(studentView, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "info").Render(ctx, w)
}

// GET request to /students/{username}/info/stream
func (s Server) getStudentViewInfoStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(username), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view info stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshStudentViewState(ctx, username); err != nil {
		s.Logger.ErrorContext(ctx, "student view info stream refresh", "err", err)
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		username+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view info stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshStudentViewState(ctx, username); err != nil {
				s.Logger.ErrorContext(ctx, "student view info stream refresh in select", "err", err)
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			student := &models.Student{}
			if err := entry.JSON(student); err != nil {
				s.Logger.ErrorContext(ctx, "student view info stream json read", "err", err)
				return
			}
			studentView := dto.NewStudentView(student)
			sse.PatchElementTempl(pages.View(studentView, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "info"))
		}
	}
}

// GET request to /students/{username}/schedule
func (s Server) getStudentViewSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get the username from the URL and get the student from the db
	username := chi.URLParam(r, "username")
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if student == nil {
		_ = pages.NotFound().Render(ctx, w)
		return
	}
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view schedule db get", "err", err)
		return
	}

	// create the student view and set the URL
	studentView := dto.NewStudentView(student)

	// get the periods for the student and make views
	periods, err := s.ReadModels.Periods.ListPeriodsForStudent(ctx, student.ID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view schedule db list periods", "err", err)
		return
	}

	personScheduleView := scheduledto.NewPersonScheduleView(student.Person, periods, true, 1)
	_ = pages.View(studentView, personScheduleView, []idto.IEPServiceView{}, "schedule").Render(ctx, w)
}

// GET request to /students/{username}/schedule/stream
func (s Server) getStudentViewScheduleStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(username), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view schedule stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshStudentViewState(ctx, username); err != nil {
		s.Logger.ErrorContext(ctx, "student view schedule stream refresh", "err", err)
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		username+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view schedule stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshStudentViewState(ctx, username); err != nil {
				s.Logger.ErrorContext(ctx, "student schedule info stream refresh in select", "err", err)
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			student := &models.Student{}
			if err := entry.JSON(student); err != nil {
				s.Logger.ErrorContext(ctx, "student view schedule stream json read", "err", err)
				return
			}
			studentView := dto.NewStudentView(student)

			// get the periods for the student and make views
			periods, err := s.ReadModels.Periods.ListPeriodsForStudent(ctx, student.ID)
			if err != nil {
				s.Logger.ErrorContext(ctx, "get student view schedule db list periods", "err", err)
				return
			}
			personScheduleView := scheduledto.NewPersonScheduleView(student.Person, periods, true, 1)
			sse.PatchElementTempl(pages.View(studentView, personScheduleView, []idto.IEPServiceView{}, "schedule"))
		}
	}
}

// GET request to /students/{username}/services
func (s Server) getStudentViewServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get the id from the URL and pull the data from the db
	username := chi.URLParam(r, "username")
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view services db get", "err", err)
		return
	}

	// create the student view and set the URL
	studentView := dto.NewStudentView(student)

	// get the list of services for the student and make views
	services, err := s.ReadModels.IEPServices.ListIEPServicesForStudent(ctx, student.ID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get student view db list services", "err", err)
	}
	serviceViews := make([]idto.IEPServiceView, len(services))
	for i, service := range services {
		serviceViews[i] = idto.NewIEPServiceView(&service)
	}
	_ = pages.View(studentView, scheduledto.PersonWithScheduleView{}, serviceViews, "services").Render(ctx, w)
}

// GET request to /students/{username}/services/stream
func (s Server) getStudentViewServicesStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(username), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view services stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshStudentViewState(ctx, username); err != nil {
		s.Logger.ErrorContext(ctx, "student view services stream refresh", "err", err)
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		username+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "student view services stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshStudentViewState(ctx, username); err != nil {
				s.Logger.ErrorContext(ctx, "student view services stream refresh in select", "err", err)
				if err.Error() == "student not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			student := &models.Student{}
			if err := entry.JSON(student); err != nil {
				s.Logger.ErrorContext(ctx, "student view services stream json read", "err", err)
				return
			}
			studentView := dto.NewStudentView(student)

			// get the list of services for the student and make views
			services, err := s.ReadModels.IEPServices.ListIEPServicesForStudent(ctx, studentView.ID)
			if err != nil {
				s.Logger.ErrorContext(ctx, "get student view db list services", "err", err)
			}
			serviceViews := make([]idto.IEPServiceView, len(services))
			for i, service := range services {
				serviceViews[i] = idto.NewIEPServiceView(&service)
			}

			sse.PatchElementTempl(pages.View(studentView, scheduledto.PersonWithScheduleView{}, serviceViews, "services"))
		}
	}
}

// GET request to /students/{username}/edit
func getStudentEdit(
	l *slog.Logger,
	studentReadModel *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Edit(dto.StudentFormView{}).Render(ctx, w)
	}
}

// GET request to /student/{username}/edit/stream
func getStudentEditStream(
	l *slog.Logger,
	vs viewstore.Store,
	subscriber MessageSubscriber,
	studentReadModel events.ReadModel,
	educatorReadModel eevents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := chi.URLParam(r, "username")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(username), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "student edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the student edit view state kv
		watcher, err := vs.Watch(
			ctx,
			username+".edit",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "student edit stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshStudentEditState(l, ctx, username, vs, studentReadModel); err != nil {
			if err.Error() == "student not found" {
				sse.PatchElementTempl(pages.NotFound())
			}
			l.ErrorContext(ctx, "student edit stream refresh in select", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				if err := refreshStudentEditState(l, ctx, username, vs, studentReadModel); err != nil {
					if err.Error() == "student not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "student edit stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				student := &models.Student{}
				if err := entry.JSON(student); err != nil {
					l.ErrorContext(ctx, "student edit stream json read", "err", err)
					return
				}
				studentFormView := dto.NewStudentFormView(student)
				caseManagers, _ := educatorReadModel.List(
					ctx,
					eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
				)

				studentFormView.CaseManagers = edto.NewEducatorSelectBoxViews(caseManagers, []string{student.CaseManager})
				sse.PatchElementTempl(pages.Edit(studentFormView))
			}
		}
	}
}

// POST request to /students/{username}/edit/validate
func (s Server) postStudentEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "student edit validate read signals", "err", err)
		return
	}
	student := dto.NewStudentModelFromView(&signals.Student)
	viewstore.PutState(ctx, s.ViewStore, username, student)
}

// POST request to /students/{username}/edit
func (s Server) postStudentEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	username := chi.URLParam(r, "username")
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post student edit read signals", "err", err)
		return
	}
	result, err := events.UpdateStudentCommandHandler(ctx, events.UpdateStudentCommand{
		StudentID:   signals.Student.ID,
		GivenName:   signals.Student.GivenName,
		ChosenName:  signals.Student.ChosenName,
		FamilyName:  signals.Student.FamilyName,
		Email:       signals.Student.Email,
		Grade:       int(signals.Student.Grade),
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post student edit update command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.InfoContext(ctx, "post student edit command handler", "skipped", result.Skipped)
		return
	}
	_, err = csevents.SyncCaseManagerForStudentCommandHandler(
		ctx,
		csevents.SyncCaseManagerForStudentCommand{
			StudentID:          signals.Student.ID,
			ProposedEducatorID: signals.Student.CaseManager,
		},
		s.EventSaver,
		s.EventRetriever,
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post student edit case manager command handler", "err", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/students/%s/info", username))
}

// POST request to /students/{username}/archive
func (s Server) postStudentArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	username := chi.URLParam(r, "username")
	student, _ := s.ReadModels.Students.GetByUsername(ctx, username)
	_, err := events.ArchiveStudentCommandHandler(ctx, events.ArchiveStudentCommand{
		StudentID: student.ID,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "archive student", "err", err)
	}
	sse := newSSE(w, r)
	sse.Redirect("/students")
}

// DELETE request to /students/{username}
func (s Server) deleteStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	username := chi.URLParam(r, "username")
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete student db get by username", "err", err)
		return
	}
	result, err := events.DeleteStudentCommandHandler(ctx, events.DeleteStudentCommand{
		StudentID: student.ID,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete student command handler", "err", err)
		return
	}
	s.Logger.InfoContext(ctx, "delete student student deleted", "id", student.ID, "event", result.EventID)
	sse := newSSE(w, r)
	sse.Redirect("/students")
}

// student helper functions

// reads the db for the given student and saves the state to a kv store for the SSE to update
func (s Server) refreshStudentViewState(ctx context.Context, username string) error {
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, username+".view", student)
}

// reads the db for the given student and saves the state to a kv store for the SSE to update
func (s Server) refreshStudentEditState(ctx context.Context, username string) error {
	student, err := s.ReadModels.Students.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, username+".edit", student)
}

// reads the db for the given student and saves the state to a kv store for the SSE to update
func refreshStudentEditState(
	_ *slog.Logger,
	ctx context.Context,
	username string,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
) error {
	student, err := studentReadModel.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, username+".edit", student)
}
