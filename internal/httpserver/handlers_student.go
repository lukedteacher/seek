package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"seek/internal/eventstore"
	"seek/internal/features/_composite/compositedto"
	"seek/internal/features/_shared/sharedmodels"
	csevents "seek/internal/features/caseload_students/events"
	edto "seek/internal/features/educators/dto"
	eevents "seek/internal/features/educators/events"
	idto "seek/internal/features/iepservices/dto"
	serviceEvents "seek/internal/features/iepservices/events"
	periodEvents "seek/internal/features/periods/events"
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
	r.Get("/students", getStudentsList(s.Logger))
	r.Get("/students/stream", getStudentsListStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Post("/students", postStudentsList(s.Logger, s.ViewStore))
	r.Post("/students/sort/{field}", postStudentsSortField(s.Logger, s.ViewStore))
	r.Get("/students/create", getStudentCreate(s.Logger))
	r.Get("/students/create/stream", getStudentCreateStream(s.Logger, s.ViewStore, *s.ReadModels.Educators))
	r.Post("/students/create/validate", postStudentCreateValidate(s.Logger, s.ViewStore))
	r.Post("/students/create", postStudentCreate(s.Logger, s.EventSaver))
	r.Get("/students/{username}", getStudentView(s.Logger))
	r.Get("/students/{username}/info", getStudentViewInfo(s.Logger))
	r.Get("/students/{username}/info/stream", getStudentViewInfoStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Get("/students/{username}/schedule", getStudentViewSchedule(s.Logger))
	r.Get("/students/{username}/schedule/stream", getStudentViewScheduleStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.Periods))
	r.Get("/students/{username}/services", getStudentViewServices(s.Logger))
	r.Get("/students/{username}/services/stream", getStudentViewServicesStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.IEPServices))
	r.Get("/students/{username}/edit", getStudentEdit(s.Logger))
	r.Get("/students/{username}/edit/stream", getStudentEditStream(s.Logger, s.ViewStore, s.Subscriber, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Post("/students/{username}/edit/validate", postStudentEditValidate(s.Logger, s.ViewStore))
	r.Post("/students/{username}/edit", postStudentEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/students/{username}/archive", postStudentArchive(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Students))
	r.Delete("/students/{username}", deleteStudent(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Students))
}

// GET request to /students
func getStudentsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		studentTableView := dto.NewStudentTableView([]models.Student{})
		_ = pages.List(pages.ListView{Table: studentTableView}).Render(ctx, w)
	}
}

// GET request to /students/stream
func getStudentsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
	educatorReadModel eevents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any students
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "students list stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches list key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			"students.list",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "students list stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		defaultFilter := make(map[string]bool, 9)
		for _, grade := range sharedmodels.GradeList {
			defaultFilter[grade.Str()] = true
		}
		listView := createListView(
			ctx,
			l,
			studentReadModel,
			"family_name",
			"ASC",
			defaultFilter,
			educatorReadModel,
		)
		sse.PatchElementTempl(pages.List(listView))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				type filterSignals struct {
					Table struct {
						Sort   map[string]int `json:"sort"`
						Filter struct {
							Grade map[string]bool `json:"grade"`
						} `json:"filter"`
					} `json:"table"`
				}
				signals, ok, err := viewstore.GetState[filterSignals](ctx, vs, "students.list")
				if err != nil {
					l.ErrorContext(ctx, "sse stream subscriber update", "err", err)
				}
				// checks if there is a view with filters different than the default
				if !ok {
					signals.Table.Filter.Grade = defaultFilter
				}
				listView := createListView(
					ctx,
					l,
					studentReadModel,
					"family_name",
					"ASC",
					signals.Table.Filter.Grade,
					educatorReadModel,
				)
				sse.PatchElementTempl(pages.List(listView))
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				signals := &struct {
					Table struct {
						Sort struct {
							Column    string
							Direction string
						} `json:"sort"`
						Filter struct {
							Grade map[string]bool `json:"grade"`
						} `json:"filter"`
					} `json:"table"`
				}{}
				if err := entry.JSON(signals); err != nil {
					l.ErrorContext(ctx, "student create stream json", "err", err)
					return
				}
				listView := createListView(
					ctx,
					l,
					studentReadModel,
					signals.Table.Sort.Column,
					signals.Table.Sort.Direction,
					signals.Table.Filter.Grade,
					educatorReadModel,
				)
				sse.PatchElementTempl(pages.List(listView))
			}
		}
	}
}

// POST request to /students
func postStudentsList(
	_ *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			Table struct {
				Sort struct {
					Column    string
					Direction string
				} `json:"sort"`
				Filter struct {
					Grade map[string]bool `json:"grade"`
				} `json:"filter"`
			} `json:"table"`
		}{}
		datastar.ReadSignals(r, signals)
		viewstore.PutState(ctx, vs, "students.list", signals)
	}
}

// POST request to /students/sort/field
func postStudentsSortField(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		field := chi.URLParam(r, "field")
		l.Debug("test", "field", field)
	}
}

// GET request to /students/create
func getStudentCreate(_ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Create(dto.StudentFormView{}).Render(ctx, w)
	}
}

// GET request to /students/create/stream
func getStudentCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel eevents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			"newstudent",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "student create stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		// populate the case managers for the select
		studentFormView := dto.NewStudentFormView(&models.Student{Grade: -1})
		caseManagers, _ := educatorReadModel.List(
			ctx,
			eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
		)

		studentFormView.CaseManagers = edto.NewEducatorSelectBoxViews(caseManagers, []string{""})
		sse.PatchElementTempl(pages.Create(studentFormView))

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
					l.ErrorContext(ctx, "student create stream json", "err", err)
					return
				}
				studentFormView := dto.NewStudentFormView(student)
				caseManagers, _ := educatorReadModel.List(
					ctx,
					eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
				)

				studentFormView.CaseManagers = edto.NewEducatorSelectBoxViews(caseManagers, []string{student.CaseManagerID})
				sse.PatchElementTempl(pages.Create(studentFormView))
			}
		}
	}
}

// POST request to /students/create/validate
func postStudentCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "student create validate signal read", "err", err)
			return
		}
		student := dto.NewStudentModelFromView(&signals.Student)
		if err := viewstore.PutState(ctx, vs, "newstudent", student); err != nil {
			l.ErrorContext(ctx, "student create validate put state", "err", err)
			return
		}
	}
}

// POST request to /student/create
func postStudentCreate(
	l *slog.Logger,
	saver eventstore.Saver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "student create signal read", "err", err)
			return
		}
		if signals.Student.Email == "" {
			sse := newSSE(w, r)
			toastError(sse, "no email provided")
			return
		}

		result, err := events.CreateStudentCommandHandler(ctx, events.CreateStudentCommand{
			MARSSID:     signals.Student.MARSSID,
			GivenName:   signals.Student.GivenName,
			ChosenName:  signals.Student.ChosenName,
			FamilyName:  signals.Student.FamilyName,
			Email:       signals.Student.Email,
			Grade:       int(signals.Student.Grade),
			Homeroom:    signals.Student.Homeroom,
			CaseManager: signals.Student.CaseManagerID,
			Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver)
		if err != nil {
			l.ErrorContext(ctx, "student create create command handler", "err", err)
			return
		}

		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/students/%s", result.EventID))
		toastSuccess(sse, "student created")
	}
}

// GET request to /students/{username}
// redirects to /students/{username}/info
func getStudentView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		http.Redirect(w, r, fmt.Sprintf("/students/%s/info", username), http.StatusFound)
	}
}

// GET request to /students/{username}/info
func getStudentViewInfo(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "info").Render(ctx, w)
	}
}

// GET request to /students/{username}/info/stream
func getStudentViewInfoStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
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
			l.ErrorContext(ctx, "student view info stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
			l.ErrorContext(ctx, "student view info stream refresh", "err", err)
			return
		}

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			username+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "student view info stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
					if err.Error() == "student not found" {
						sse.PatchElementTempl(pages.NotFound())
						return
					}
					l.ErrorContext(ctx, "student view info stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				student := &models.Student{}
				if err := entry.JSON(student); err != nil {
					l.ErrorContext(ctx, "student view info stream json read", "err", err)
					return
				}
				studentView := dto.StudentView{}
				if student.CaseManagerID != "" {
					caseManager, err := educatorReadModel.GetByID(ctx, student.CaseManagerID)
					studentView = dto.NewStudentView(student, caseManager)
					if err != nil {
						l.ErrorContext(ctx, "student view info db case manager get", "err", err)
						return
					}
				} else {
					studentView = dto.NewStudentView(student, nil)
				}
				sse.PatchElementTempl(pages.View(studentView, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "info"))
			}
		}
	}
}

// GET request to /students/{username}/schedule
func getStudentViewSchedule(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "schedule").Render(ctx, w)
	}
}

// GET request to /students/{username}/schedule/stream
func getStudentViewScheduleStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
	periodReadModel periodEvents.ReadModel,
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
			l.ErrorContext(ctx, "student view schedule stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
			l.ErrorContext(ctx, "student view schedule stream refresh", "err", err)
			return
		}

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			username+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "student view schedule stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
					if err.Error() == "student not found" {
						sse.PatchElementTempl(pages.NotFound())
						return
					}
					l.ErrorContext(ctx, "student schedule info stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				student := &models.Student{}
				if err := entry.JSON(student); err != nil {
					l.ErrorContext(ctx, "student view schedule stream json read", "err", err)
					return
				}
				studentView := dto.NewStudentView(student, nil)

				// get the periods for the student and make views
				periods, err := periodReadModel.ListPeriodsForStudent(ctx, student.ID)
				if err != nil {
					l.ErrorContext(ctx, "get student view schedule db list periods", "err", err)
					return
				}
				personScheduleView := scheduledto.NewPersonScheduleView(student.Person, periods, true, 1)
				sse.PatchElementTempl(pages.View(studentView, personScheduleView, []idto.IEPServiceView{}, "schedule"))
			}
		}
	}
}

// GET request to /students/{username}/services
func getStudentViewServices(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.IEPServiceView{}, "services").Render(ctx, w)
	}
}

// GET request to /students/{username}/services/stream
func getStudentViewServicesStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
	iepServiceReadModel serviceEvents.ReadModel,
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
			l.ErrorContext(ctx, "student view services stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
			l.ErrorContext(ctx, "student view services stream refresh", "err", err)
			return
		}

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			username+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "student view services stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshStudentViewState(ctx, l, username, vs, studentReadModel); err != nil {
					l.ErrorContext(ctx, "student view services stream refresh in select", "err", err)
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
					l.ErrorContext(ctx, "student view services stream json read", "err", err)
					return
				}
				studentView := dto.NewStudentView(student, nil)

				// get the list of services for the student and make views
				services, err := iepServiceReadModel.ListIEPServicesForStudent(ctx, studentView.ID)
				if err != nil {
					l.ErrorContext(ctx, "get student view db list services", "err", err)
				}
				serviceViews := make([]idto.IEPServiceView, len(services))
				for i, service := range services {
					serviceViews[i] = idto.NewIEPServiceView(&service)
				}

				sse.PatchElementTempl(pages.View(studentView, scheduledto.PersonWithScheduleView{}, serviceViews, "services"))
			}
		}
	}
}

// GET request to /students/{username}/edit
func getStudentEdit(
	_ *slog.Logger,
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

				// list current case managers to populate form
				caseManagers, _ := educatorReadModel.List(
					ctx,
					eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
				)
				studentFormView.CaseManagers = edto.NewEducatorSelectBoxViews(caseManagers, []string{student.CaseManagerID})

				// patch data to page
				sse.PatchElementTempl(pages.Edit(studentFormView))
			}
		}
	}
}

// POST request to /students/{username}/edit/validate
func postStudentEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := chi.URLParam(r, "username")
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "student edit validate read signals", "err", err)
			return
		}
		student := dto.NewStudentModelFromView(&signals.Student)
		viewstore.PutState(ctx, vs, username, student)
	}
}

// POST request to /students/{username}/edit
func postStudentEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		username := chi.URLParam(r, "username")
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post student edit read signals", "err", err)
			return
		}
		result, err := events.UpdateStudentCommandHandler(
			ctx,
			events.UpdateStudentCommand{
				StudentID:   signals.Student.ID,
				MARSSID:     signals.Student.MARSSID,
				GivenName:   signals.Student.GivenName,
				ChosenName:  signals.Student.ChosenName,
				FamilyName:  signals.Student.FamilyName,
				Email:       signals.Student.Email,
				Grade:       int(signals.Student.Grade),
				Homeroom:    signals.Student.Homeroom,
				CaseManager: signals.Student.CaseManagerID,
				Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
			},
			saver,
			retriever,
		)
		if err != nil {
			l.ErrorContext(ctx, "post student edit update command handler", "err", err)
			return
		}
		if result.Skipped == true {
			l.InfoContext(ctx, "post student edit command handler", "skipped", result.Skipped)
			return
		}
		if signals.Student.CaseManagerID != "" {
			_, err = csevents.SyncCaseManagerForStudentCommandHandler(
				ctx,
				csevents.SyncCaseManagerForStudentCommand{
					StudentID:          signals.Student.ID,
					ProposedEducatorID: signals.Student.CaseManagerID,
				},
				saver,
				retriever,
			)
			if err != nil {
				l.ErrorContext(ctx, "post student edit case manager command handler", "err", err)
				return
			}
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/students/%s/info", username))
	}
}

// POST request to /students/{username}/archive
func postStudentArchive(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	studentReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		username := chi.URLParam(r, "username")
		student, _ := studentReadModel.GetByUsername(ctx, username)
		_, err := events.ArchiveStudentCommandHandler(ctx, events.ArchiveStudentCommand{
			StudentID: student.ID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "archive student", "err", err)
		}
		sse := newSSE(w, r)
		sse.Redirect("/students")
	}
}

// DELETE request to /students/{username}
func deleteStudent(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	studentReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		username := chi.URLParam(r, "username")
		student, err := studentReadModel.GetByUsername(ctx, username)
		if err != nil {
			l.ErrorContext(ctx, "delete student db get by username", "err", err)
			return
		}
		result, err := events.DeleteStudentCommandHandler(ctx, events.DeleteStudentCommand{
			StudentID: student.ID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete student command handler", "err", err)
			return
		}
		l.InfoContext(ctx, "delete student student deleted", "id", student.ID, "event", result.EventID)
		sse := newSSE(w, r)
		sse.Redirect("/students")
	}
}

// student helper functions

func createListView(
	ctx context.Context,
	l *slog.Logger,
	studentReadModel events.ReadModel,
	sortCol,
	sortDir string,
	studentFilter map[string]bool,
	educatorReadModel eevents.ReadModel,
) pages.ListView {
	// get students data from db
	// temp := temp(studentFilter)
	students, err := studentReadModel.List(ctx, events.WithSort(sortCol, sortDir))
	if err != nil {
		l.ErrorContext(ctx, "create students view", "err", err)
		return pages.ListView{}
	}
	// create the views
	studentWithDataViews := make([]compositedto.StudentWithData, len(students))
	for i, student := range students {
		studentWithDataViews[i] = compositedto.StudentWithData{
			Student: student,
		}
		if student.CaseManagerID != "" {
			caseManager, err := educatorReadModel.GetByID(ctx, student.CaseManagerID)
			if err != nil {
				l.ErrorContext(ctx, "student list sse get case manager from db", "err", err)
			}
			if caseManager != nil {
				studentWithDataViews[i].CaseManager = *caseManager
			}
		}
	}

	// create the table view
	studentTableView := compositedto.NewStudentWithDataTableView(studentWithDataViews)
	studentTableView.Sort.Column = sortCol
	studentTableView.Sort.Direction = sortDir

	return pages.ListView{
		Table:        studentTableView,
		FilterGrades: studentFilter,
	}
}

func temp(m map[string]bool) []int {
	var result []int
	for k, v := range m {
		if v {
			if i, err := strconv.Atoi(k); err == nil {
				result = append(result, i)
			}
		}
	}
	return result
}

// reads the db for the given student and saves the state to a kv store for the SSE to update
func refreshStudentViewState(
	ctx context.Context,
	_ *slog.Logger,
	username string,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
) error {
	student, err := studentReadModel.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, username+".view", student)
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
