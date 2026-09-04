package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"seek/internal/eventstore"
	"seek/internal/features/_composite/compositedto"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	eevents "seek/internal/features/educators/events"
	periodEvents "seek/internal/features/periods/events"
	scheduledto "seek/internal/features/schedules/dto"
	idto "seek/internal/features/services/dto"
	serviceEvents "seek/internal/features/services/events"
	"seek/internal/features/students/dto"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
	"seek/internal/features/students/pages"
	"seek/internal/ui/core/coreblocks/toasts"
	"seek/internal/viewstore"
	"seek/pkg/templui/components/toast"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students", getStudentsList(s.Logger))
	r.Get("/students/stream", getStudentsListStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Post("/students", postStudentsList(s.Logger, s.ViewStore))
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
	r.Get("/students/{username}/services/stream", getStudentViewServicesStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Students, *s.ReadModels.Services))
	r.Get("/students/{username}/edit", getStudentEdit(s.Logger))
	r.Get("/students/{username}/edit/stream", getStudentEditStream(s.Logger, s.ViewStore, s.Subscriber, *s.ReadModels.Students, *s.ReadModels.Educators))
	r.Post("/students/{username}/edit/validate", postStudentEditValidate(s.Logger, s.ViewStore))
	r.Post("/students/{username}/edit", postStudentEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/students/{username}/archive", postStudentArchive(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Students))
	r.Delete("/students/{username}", deleteStudent(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Students))
	r.Get("/students/csv", getStudentsCSV(s.Logger, *s.ReadModels.Students))
	r.Post("/students/csv", postStudentsCSV(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Students))
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
		user := currentUser(r)

		// subscribes to the channel which publishes changes to any students
		notifier := NewMessageNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(ctx context.Context, data []byte) {
			notifier.Notify(data)
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
			user.Username+".students.list",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "students list stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		defaultGradeFilter := make(map[string]bool, 9)
		for _, grade := range sharedmodels.GradeList {
			defaultGradeFilter[grade.String()] = true
		}
		defaultPlanTypeFilter := make(map[string]bool, 4)
		for _, planType := range sharedmodels.PlanTypeList {
			defaultPlanTypeFilter[planType.String()] = true
		}
		listView := createListView(
			ctx,
			l,
			studentReadModel,
			"family_name",
			"ASC",
			dto.StudentFilter{
				Grade:    defaultGradeFilter,
				PlanType: defaultPlanTypeFilter,
				Search:   "",
			},
			educatorReadModel,
		)
		sse.PatchElementTempl(pages.List(listView))

		for {
			select {
			case <-ctx.Done():
				return
			case data := <-notifier.Signal(): // triggers when the read model publishes
				var msg map[string]string
				if err := json.Unmarshal(data, &msg); err != nil {
					l.ErrorContext(ctx, "parse signal data", "err", err)
					continue
				}
				toastMsg := fmt.Sprintf("Updated: %s", msg["studentID"]) // example

				type tableSignals struct {
					Table dto.StudentTableState `json:"table"`
				}
				signals, ok, err := viewstore.GetState[tableSignals](ctx, vs, user.Username+".students.list")
				if err != nil {
					l.ErrorContext(ctx, "student create stream subscriber update", "err", err)
				}
				if !ok {
					signals.Table.Sort.Column = "family_name"
					signals.Table.Sort.Direction = "ASC"
				}
				listView := createListView(
					ctx,
					l,
					studentReadModel,
					signals.Table.Sort.Column,
					signals.Table.Sort.Direction,
					signals.Table.Filter,
					educatorReadModel,
				)
				sse.PatchElementTempl(pages.List(listView))
				sse.PatchElementTempl(toasts.ToastContainer(toast.VariantInfo, toastMsg))
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				signals := &struct {
					Table dto.StudentTableState `json:"table"`
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
					signals.Table.Filter,
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
		user := currentUser(r)
		signals := &struct {
			Table dto.StudentTableState `json:"table"`
		}{}
		datastar.ReadSignals(r, signals)
		viewstore.PutState(ctx, vs, user.Username+".students.list", signals)
	}
}

// GET request to /students/create
func getStudentCreate(_ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Create(dto.StudentFormView{FormType: "create"}).Render(ctx, w)
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
		user := currentUser(r)
		sse := newSSE(w, r)

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			user.Username+".students.create",
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
		studentFormView := dto.NewStudentFormView("create", &models.Student{Grade: -1})
		caseManagers, _ := educatorReadModel.List(
			ctx,
			eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
		)
		studentFormView.PlanTypeOptions = dto.NewSelectPlanTypeOptions(sharedmodels.PlanTypeList, sharedmodels.PlanTypeNone)
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
				studentFormView := dto.NewStudentFormView("create", student)
				caseManagers, _ := educatorReadModel.List(
					ctx,
					eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
				)
				studentFormView.PlanTypeOptions = dto.NewSelectPlanTypeOptions(sharedmodels.PlanTypeList, student.PlanType)
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
		user := currentUser(r)
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "student create validate signal read", "err", err)
			return
		}
		student := dto.NewStudentModelFromView(&signals.Student)
		if err := viewstore.PutState(ctx, vs, user.Username+".students.create", student); err != nil {
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
		student := events.StudentState{
			ID:         signals.Student.ID,
			MARSSID:    signals.Student.MARSSID,
			GivenName:  signals.Student.GivenName,
			ChosenName: signals.Student.ChosenName,
			FamilyName: signals.Student.FamilyName,
			Email:      signals.Student.Email,
			Username:   signals.Student.Username,
			Grade:      int(signals.Student.Grade),
			HomeroomID: signals.Student.HomeroomID,
			PlanType:   int(signals.Student.PlanType),
		}
		cmd := events.CreateStudentCommand{
			StudentState: student,
			Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.CreateStudentCommandHandler(ctx, cmd, saver)
		if err != nil {
			l.ErrorContext(ctx, "student create create command handler", "err", err)
			return
		}

		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/students/%s", result.Student.Username))
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
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.ServiceView{}, "info").Render(ctx, w)
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

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
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
				sse.PatchElementTempl(pages.View(studentView, scheduledto.PersonWithScheduleView{}, []idto.ServiceView{}, "info"))
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
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.ServiceView{}, "schedule").Render(ctx, w)
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
				personScheduleView := scheduledto.NewPersonScheduleView(student.ID, student.Person, periods, true, 1)
				sse.PatchElementTempl(pages.View(studentView, personScheduleView, []idto.ServiceView{}, "schedule"))
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
		_ = pages.View(dto.StudentView{}, scheduledto.PersonWithScheduleView{}, []idto.ServiceView{}, "services").Render(ctx, w)
	}
}

// GET request to /students/{username}/services/stream
func getStudentViewServicesStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	studentReadModel events.ReadModel,
	serviceReadModel serviceEvents.ReadModel,
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
				services, err := serviceReadModel.ListServicesForIEP(ctx, studentView.ID)
				if err != nil {
					l.ErrorContext(ctx, "get student view db list services", "err", err)
				}
				serviceViews := make([]idto.ServiceView, len(services))
				for i, service := range services {
					serviceViews[i] = idto.NewServiceView(&service)
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

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
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
				studentFormView := dto.NewStudentFormView("edit", student)

				// list current case managers to populate form
				caseManagers, _ := educatorReadModel.List(
					ctx,
					eevents.FilterByRole(sharedmodels.EducatorRoleCaseManager),
				)
				studentFormView.PlanTypeOptions = dto.NewSelectPlanTypeOptions(sharedmodels.PlanTypeList, student.PlanType)
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
		if signals.Student.ID == "" {
			l.ErrorContext(ctx, "student edit validate no student id")
			return
		}
		student := dto.NewStudentModelFromView(&signals.Student)
		viewstore.PutState(ctx, vs, username+".edit", student)
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
		signals := &struct {
			Student dto.StudentView `json:"student"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post student edit read signals", "err", err)
			return
		}
		student := events.StudentState{
			ID:         signals.Student.ID,
			MARSSID:    signals.Student.MARSSID,
			GivenName:  signals.Student.GivenName,
			ChosenName: signals.Student.ChosenName,
			FamilyName: signals.Student.FamilyName,
			Email:      signals.Student.Email,
			Grade:      int(signals.Student.Grade),
			Username:   signals.Student.Username,
			HomeroomID: signals.Student.HomeroomID,
			PlanType:   int(signals.Student.PlanType),
		}
		cmd := events.UpdateStudentCommand{
			StudentState: student,
			Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateStudentCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post student edit update command handler", "err", err)
			return
		}
		if result.Skipped == true {
			l.InfoContext(ctx, "post student edit command handler", "skipped", result.Skipped)
			return
		}
		// if signals.Student.CaseManagerID != "" {
		// 	_, err = csevents.SyncCaseManagerForStudentCommandHandler(
		// 		ctx,
		// 		csevents.SyncCaseManagerForStudentCommand{
		// 			StudentID:          signals.Student.ID,
		// 			ProposedEducatorID: signals.Student.CaseManagerID,
		// 		},
		// 		saver,
		// 		retriever,
		// 	)
		// 	if err != nil {
		// 		l.ErrorContext(ctx, "post student edit case manager command handler", "err", err)
		// 		return
		// 	}
		// }
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/students/%s/info", result.Student.Username))
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
		cmd := events.DeleteStudentCommand{
			StudentID: student.ID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.DeleteStudentCommandHandler(ctx, cmd, saver, retriever)
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
	filters dto.StudentFilter,
	educatorReadModel eevents.ReadModel,
) pages.ListView {
	// get students data from db
	gradeFilter := buildFilterMap(filters.Grade)
	planTypeFilter := buildFilterMap(filters.PlanType)
	students, err := studentReadModel.List(
		ctx,
		events.WithSort(sortCol, sortDir),
		events.WithGradeFilter(gradeFilter),
		events.WithPlanFilter(planTypeFilter),
		events.WithSearchFilter(filters.Search),
	)
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
	studentTableView := compositedto.NewStudentWithDataTableView(studentWithDataViews, shareddto.TableSort{
		Column:    sortCol,
		Direction: sortDir,
	})

	return pages.ListView{
		Table: studentTableView,
		Filters: dto.StudentFilter{
			Grade:    filters.Grade,
			PlanType: filters.PlanType,
		},
	}
}

func buildFilterMap(m map[string]bool) []int {
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

// GET request to /students/csv
func getStudentsCSV(
	l *slog.Logger,
	studentReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, err := os.OpenFile("/home/lukeout/seek/students.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			http.Error(w, "failed to open csv file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		csvStudents := []models.Student{}
		if err := gocsv.UnmarshalFile(file, &csvStudents); err != nil {
			http.Error(w, "failed to parse csv: "+err.Error(), http.StatusBadRequest)
			return
		}

		dbStudents, err := studentReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "get students csv", "err", err)
		}

		diffs := models.CompareStudents(dbStudents, csvStudents)

		// render view
		view := dto.NewStudentsDiffTableView(diffs)
		pages.ReadStudentCSV(view).Render(ctx, w)
	}
}

// POST request to /students/csv
func postStudentsCSV(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	studentReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, err := os.OpenFile("/home/lukeout/seek/students.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			http.Error(w, "failed to open csv file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		csvStudents := []models.Student{}
		if err := gocsv.UnmarshalFile(file, &csvStudents); err != nil {
			http.Error(w, "failed to parse csv: "+err.Error(), http.StatusBadRequest)
			return
		}

		dbStudents, err := studentReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "post students csv list", "err", err)
		}

		diffs := models.CompareStudents(dbStudents, csvStudents)

		for _, diff := range diffs {
			switch diff.Status {
			case sharedmodels.DiffSame:
				continue
			case sharedmodels.DiffNew:
				cmd := events.CreateStudentCommand{
					StudentState: events.StudentState{
						GivenName:  diff.New.GivenName,
						ChosenName: diff.New.ChosenName,
						FamilyName: diff.New.FamilyName,
						Email:      diff.New.Email,
						Username:   diff.New.Username,
						Grade:      int(diff.New.Grade),
					},
				}
				_, err := events.CreateStudentCommandHandler(
					ctx,
					cmd,
					saver,
				)
				if err != nil {
					l.ErrorContext(ctx, "post students csv create", "err", err)
				}
			case sharedmodels.DiffUpdated:
				cmd := events.UpdateStudentCommand{
					StudentState: events.StudentState{
						ID:         diff.Old.ID,
						GivenName:  diff.New.GivenName,
						ChosenName: diff.New.ChosenName,
						FamilyName: diff.New.FamilyName,
						Email:      diff.New.Email,
						Username:   diff.New.Username,
						Grade:      int(diff.New.Grade),
					},
				}
				_, err := events.UpdateStudentCommandHandler(
					ctx,
					cmd,
					saver,
					retriever,
				)
				if err != nil {
					l.ErrorContext(ctx, "post students csv update", "err", err)
				}
			case sharedmodels.DiffAbsent:
				cmd := events.ArchiveStudentCommand{
					StudentID: diff.Old.ID,
				}
				_, err := events.ArchiveStudentCommandHandler(
					ctx,
					cmd,
					saver,
					retriever,
				)
				if err != nil {
					l.ErrorContext(ctx, "post students csv archive", "err", err)
				}
			}
		}

		sse := newSSE(w, r)
		sse.Redirect("/students")
	}
}
