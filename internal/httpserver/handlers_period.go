package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/_composite/compositedto"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	educatorEvents "seek/internal/features/educators/events"
	epevents "seek/internal/features/educators_periods/events"
	"seek/internal/features/periods/blocks"
	"seek/internal/features/periods/dto"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	"seek/internal/features/periods/pages"
	scheduleDTO "seek/internal/features/schedules/dto"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	spevents "seek/internal/features/students_periods/events"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	// period list
	r.Get("/periods", getPeriodsList(s.Logger))
	r.Get("/periods/stream", getPeriodsListStream(s.Logger, s.Subscriber, s.ReadModels.Periods, s.ReadModels.Educators, s.ReadModels.Students))
	// create periods
	r.Get("/periods/create", getPeriodCreate(s.Logger))
	r.Get("/periods/create/stream", getPeriodCreateStream(s.Logger, s.ViewStore, s.ReadModels.Periods, s.ReadModels.Students, s.ReadModels.Educators))
	r.Query("/periods/create/validate", queryPeriodFormValidate(s.Logger, s.ViewStore))
	r.Query("/periods/create/validate/{field}", queryPeriodCreateValidateField(s.Logger, s.ViewStore))
	r.Query("/periods/create/{field}/{value}", queryPeriodFormField(s.Logger, s.ViewStore))
	r.Post("/periods/create", postPeriodCreate(s.Logger, s.EventSaver, s.EventRetriever))
	// view period
	r.Get("/periods/{id}", getPeriodView(s.Logger))
	r.Get("/periods/{id}/stream", getPeriodViewStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Periods, s.ReadModels.Educators, s.ReadModels.Students))
	// edit periods
	r.Get("/periods/{id}/edit", getPeriodEdit(s.Logger))
	r.Get("/periods/{id}/edit/stream", getPeriodEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Periods, s.ReadModels.Students, s.ReadModels.Educators))
	r.Query("/periods/{id}/edit/validate", queryPeriodFormValidate(s.Logger, s.ViewStore))
	r.Query("/periods/{id}/edit/validate/{field}", queryPeriodCreateValidateField(s.Logger, s.ViewStore))
	r.Query("/periods/{id}/edit/{field}/{value}", queryPeriodFormField(s.Logger, s.ViewStore))
	r.Post("/periods/{id}/edit", postPeriodEdit(s.Logger, s.EventSaver, s.EventRetriever))
	// other period functions
	r.Post("/periods/{id}/archive", postPeriodArchive(s.Logger, s.EventSaver, s.EventRetriever))
	r.Delete("/periods/{id}", deletePeriod(s.Logger, s.EventSaver, s.EventRetriever))
}

// GET request to /periods
// renders an empty template
// SSE will populate data
func getPeriodsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewPeriodTableView([]models.Period{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /periods/stream
// populates data and keeps it updated with changes pushed from server
func getPeriodsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	periodReadModel *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any periods
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "periods list stream subscribe", "err", err)
			return
		}
		defer sub.Close()
		view := createPeriodsListView(ctx, l, *periodReadModel, *educatorReadModel, *studentReadModel)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				// for now just reloads the page
				// consider adding a vi

				view := createPeriodsListView(ctx, l, *periodReadModel, *educatorReadModel, *studentReadModel)
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /periods/create
func getPeriodCreate(_ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		props := pages.PeriodFormPageProps{
			FormProps: blocks.PeriodFormProps{
				FormType: sharedmodels.FormTypeCreate,
			},
		}
		_ = pages.PeriodForm(props).Render(ctx, w)
	}
}

// GET request to /periods/create/stream
func getPeriodCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	periodReadModel *events.ReadModel,
	studentReadModel *studentEvents.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// watch for view store changes
		key, err := getPeriodViewstoreKey(sharedmodels.FormTypeCreate, user.Username, "")
		if err != nil {
			l.ErrorContext(ctx, "gpcs vs key", "err", err)
			return
		}
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "gpcs watch", "err", err)
			return
		}
		defer watcher.Stop()

		if err := initializePeriodCreateState(ctx, l, vs, user.Username); err != nil {
			l.ErrorContext(ctx, "gpcs init vs", "err", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.PeriodFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.Error("json decode", "err", err)
					return
				}
				model := dto.NewPeriodModelFromView(&signals.Period)
				educators := listEducatorsByIDs(ctx, l, educatorReadModel, signals.Period.EducatorIDs)
				students := listStudentsByIDs(ctx, l, studentReadModel, signals.Period.StudentIDs)
				scheduleViews := buildScheduleViews(
					ctx,
					l,
					model,
					nil,
					periodReadModel,
					studentReadModel,
				)
				props := pages.PeriodFormPageProps{
					FormProps: blocks.PeriodFormProps{
						FormType:          sharedmodels.FormTypeCreate,
						Period:            signals.Period,
						EducatorSelect:    createEducatorSelectView(ctx, l, educatorReadModel, &signals.EducatorSelect.Filter, signals.Period.EducatorIDs),
						SelectedEducators: educatorDTO.NewViews(educators),
						StudentSelect:     createStudentSelectView(ctx, l, studentReadModel, &signals.StudentSelect.Filter, signals.Period.StudentIDs),
						SelectedStudents:  studentDTO.NewViews(students),
					},
					Schedules: scheduleViews,
				}
				for _, s := range props.FormProps.StudentSelect.Options {
					l.Debug("***", "sid", s.ID, "n", s.GivenName)
				}
				sse.PatchElementTempl(pages.PeriodForm(props))
			}
		}
	}
}

// QUERY request to /periods/create/validate
// reads datastar signals from the form and saves them to the view store.
// this allows the SSE stream to detect changes and refresh the form preview.
func queryPeriodFormValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &dto.PeriodFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qpfv signals", "err", err.Error())
			return
		}
		key, err := getPeriodViewstoreKey(signals.FormType, user.Username, signals.Period.ID)
		if err != nil {
			l.ErrorContext(ctx, "qpfv get vs key", "err", err, "form type", signals.FormType.Word(), "username", user.Username, "period id", signals.Period.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "gpfv vs", "err", err)
			return
		}
	}
}

func queryPeriodCreateValidateField(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &dto.PeriodFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "pcvf signal read", "error", err)
			return
		}
		field := chi.URLParam(r, "field")

		switch field {
		case "start_time":
			// update start time → recalculate end time using current duration
			signals.Period.UpdateStartTime(signals.Period.StartTime)

		case "end_time":
			// update end time → recalculate duration using current start time
			signals.Period.UpdateEndTime(signals.Period.EndTime)

		case "duration":
			// update duration → recalculate end time using current start time
			signals.Period.UpdateDuration(signals.Period.Duration)

		default:
			l.WarnContext(ctx, "unknown validation field", "field", field)
			http.Error(w, "unknown field", http.StatusBadRequest)
			return
		}
		key, err := getPeriodViewstoreKey(signals.FormType, user.Username, signals.Period.ID)
		if err != nil {
			l.ErrorContext(ctx, "qpcvf vs key", "err", err)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
			return
		}
	}
}

func queryPeriodFormField(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		field := chi.URLParam(r, "field")
		value := chi.URLParam(r, "value")
		signals := &dto.PeriodFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qpff signals", "err", err)
			return
		}
		switch field {
		case "servicetype":
			serviceType, err := sharedmodels.ParseServiceType(value)
			if err != nil {
				l.ErrorContext(ctx, "qpff st parse", "err", err)
			}
			signals.Period.ServiceType = serviceType
		case "educators":
			signals.Period.EducatorIDs = toggleID(signals.Period.EducatorIDs, value)
		case "days":
			day, err := sharedmodels.ParseDay(value)
			if err != nil {
				l.ErrorContext(ctx, "qpff day parse", "err", err)
				return
			}
			signals.Period.DaysBitmask.ToggleDay(day)
		case "students":
			signals.Period.StudentIDs = toggleID(signals.Period.StudentIDs, value)
		default:
			l.ErrorContext(ctx, "qpff", "invalid field in form", field)
		}
		key, err := getPeriodViewstoreKey(signals.FormType, user.Username, signals.Period.ID)
		if err != nil {
			l.ErrorContext(ctx, "qpff vs key", "err", err, "form type", signals.FormType, "username", user.Username, "hid", signals.Period.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qpff vs", "error", err)
			return
		}
	}
}

// POST request to /periods/create
// reads the form signals, creates a new period, syncs students and educators,
// then redirects to the period view page.
func postPeriodCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)

		signals := &struct {
			Period dto.PeriodFormView `json:"period"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post period create signals", "err", err)
			return
		}

		// create the period
		cmd := events.CreatePeriodCommand{
			Title:       signals.Period.Title,
			ServiceType: signals.Period.ServiceType,
			StartTime:   signals.Period.StartTime,
			Duration:    signals.Period.Duration,
			DaysBitmask: signals.Period.Days.ToBitmask(),
			Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.CreatePeriodCommandHandler(ctx, cmd, saver)
		if err != nil {
			l.ErrorContext(ctx, "post period create command handler", "err", err)
			return
		}

		periodID := result.EventID

		// sync educators
		secmd := epevents.SyncEducatorsInPeriodCommand{
			PeriodID:            periodID,
			ProposedEducatorIDs: signals.Period.EducatorIDs,
		}
		if _, err := epevents.SyncEducatorsInPeriodCommandHandler(ctx, secmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period create sync educators", "err", err)
		}

		// sync students
		spcmd := spevents.SyncStudentsInPeriodCommand{
			PeriodID:           periodID,
			ProposedStudentIDs: signals.Period.StudentIDs,
		}
		if _, err := spevents.SyncStudentsInPeriodCommandHandler(ctx, spcmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period create sync students", "err", err)
		}

		// redirect to the new period view
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/periods/%s", periodID))
	}
}

// GET request to /periods/{id}
func getPeriodView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// minimal empty view
		view := dto.PeriodView{}
		_ = pages.View(view).Render(ctx, w)
	}
}
func getPeriodViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	periodReadModel *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		periodID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "get period view stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			periodID+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "get period view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshPeriodViewState(ctx, l, periodID, periodReadModel, vs); err != nil {
			l.ErrorContext(ctx, "get period view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshPeriodViewState(ctx, l, periodID, periodReadModel, vs); err != nil {
					// the period was deleted or archived
					if err.Error() == "period not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "get period view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				period := &models.Period{}
				if err := entry.JSON(period); err != nil {
					l.ErrorContext(ctx, "get period view stream json", "err", err)
					return
				}
				view := dto.NewPeriodView(period)
				for i := range period.EducatorIDs {
					educator, _ := educatorReadModel.GetByID(ctx, period.EducatorIDs[i])
					educatorView := educatorDTO.NewView(educator)
					view.Educators = append(view.Educators, educatorView)
				}
				for i := range period.StudentIDs {
					student, _ := studentReadModel.GetByID(ctx, period.StudentIDs[i])
					studentView := studentDTO.NewView(student)
					view.Students = append(view.Students, studentView)
				}
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

func getPeriodEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// creates a minimal view
		props := pages.PeriodFormPageProps{
			FormProps: blocks.PeriodFormProps{
				FormType: sharedmodels.FormTypeEdit,
			},
		}
		_ = pages.PeriodForm(props).Render(ctx, w)
	}
}

// GET request to /periods/{id}/edit/stream
// establishes an SSE connection that updates the edit form when the period or view state changes.
func getPeriodEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	periodReadModel *events.ReadModel,
	studentReadModel *studentEvents.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		periodID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// subscribe to event store changes
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "period edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		key, err := getPeriodViewstoreKey(sharedmodels.FormTypeEdit, "", periodID)
		if err != nil {
			l.ErrorContext(ctx, "gpes vs key", "err", err)
		}

		// subscribe to the kv store for changes to the edit view state
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "edit view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		// check if the kv store has an edit view already created
		// aka someone else is editing the period
		// if not, populate the view with data from the db
		_, ok, err := vs.Get(ctx, key)
		if !ok {
			if err := refreshPeriodEditState(ctx, l, vs, periodReadModel, periodID); err != nil {
				if err.Error() == "period not found" {
					sse.PatchElementTempl(pages.NotFound())
				} else {
					l.ErrorContext(ctx, "refresh period view state", "err", err)
				}
				return
			}
		}
		if err != nil {
			l.ErrorContext(ctx, "period edit stream", "vs get err", err)
		}

		for {
			select {
			case <-ctx.Done():
				return

			case <-notifier.Signal():
				// period changed via event – refresh the view state and re‑render
				if err := refreshPeriodEditState(ctx, l, vs, periodReadModel, periodID); err != nil {
					if err.Error() == "period not found" {
						sse.PatchElementTempl(pages.NotFound())
					} else {
						l.ErrorContext(ctx, "refresh period view state", "err", err)
					}
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.PeriodFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.Error("period edit stream json", "err", err)
					return
				}
				model := dto.NewPeriodModelFromView(&signals.Period)
				educators := listEducatorsByIDs(ctx, l, educatorReadModel, signals.Period.EducatorIDs)
				students := listStudentsByIDs(ctx, l, studentReadModel, signals.Period.StudentIDs)
				scheduleViews := buildScheduleViews(
					ctx,
					l,
					model,
					nil,
					periodReadModel,
					studentReadModel,
				)
				props := pages.PeriodFormPageProps{
					FormProps: blocks.PeriodFormProps{
						FormType:          sharedmodels.FormTypeEdit,
						Period:            signals.Period,
						EducatorSelect:    createEducatorSelectView(ctx, l, educatorReadModel, &signals.EducatorSelect.Filter, signals.Period.EducatorIDs),
						SelectedEducators: educatorDTO.NewViews(educators),
						StudentSelect:     createStudentSelectView(ctx, l, studentReadModel, &signals.StudentSelect.Filter, signals.Period.StudentIDs),
						SelectedStudents:  studentDTO.NewViews(students),
					},
					Schedules: scheduleViews,
				}
				sse.PatchElementTempl(pages.PeriodForm(props))
			}
		}
	}
}

// POST request to /periods/{id}/edit
// reads the form signals, updates the period, syncs educators and students,
// then redirects to the period view page.
func postPeriodEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &dto.PeriodFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post period edit signals", "err", err)
			return
		}
		periodID := chi.URLParam(r, "id")

		// update the period itself
		cmd := events.UpdatePeriodCommand{
			ID:          periodID,
			Title:       signals.Period.Title,
			ServiceType: signals.Period.ServiceType,
			StartTime:   signals.Period.StartTime,
			Duration:    signals.Period.Duration,
			DaysBitmask: signals.Period.DaysBitmask,
			Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdatePeriodCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post period edit command handler", "err", err, "pid", periodID)
			return
		}
		if result.Skipped {
			l.InfoContext(ctx, "post period edit command handler", "skipped", result.Skipped)
		}

		// sync educators
		secmd := epevents.SyncEducatorsInPeriodCommand{
			PeriodID:            periodID,
			ProposedEducatorIDs: signals.Period.EducatorIDs,
		}
		if _, err := epevents.SyncEducatorsInPeriodCommandHandler(ctx, secmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period edit sync educators", "err", err)
		}

		// sync students
		spcmd := spevents.SyncStudentsInPeriodCommand{
			PeriodID:           periodID,
			ProposedStudentIDs: signals.Period.StudentIDs,
		}
		if _, err := spevents.SyncStudentsInPeriodCommandHandler(ctx, spcmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period edit sync students", "err", err)
		}

		// redirect to the period view
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/periods/%s", periodID))
	}
}

// POST request to /periods/{id}/archive
func postPeriodArchive(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		periodID := chi.URLParam(r, "id")
		_, err := events.ArchivePeriodCommandHandler(ctx, events.ArchivePeriodCommand{
			PeriodID: periodID,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "archive period command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/periods")
	}
}

// DELETE request to /periods/{id}
func deletePeriod(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		periodID := chi.URLParam(r, "id")
		_, err := events.DeletePeriodCommandHandler(ctx, events.DeletePeriodCommand{
			PeriodID: periodID,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete period command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/periods")
	}
}

// gets period from the db and saves it to a kv store
func refreshPeriodViewState(
	ctx context.Context,
	_ *slog.Logger,
	periodID string,
	periods *events.ReadModel,
	vs viewstore.Store,
) error {
	model, err := periods.GetWithIDs(ctx, periodID)
	if err != nil {
		return err
	}
	key := model.ID + ".view"
	return viewstore.PutState(ctx, vs, key, model)
}

func initializePeriodCreateState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	username string,
) error {
	signals := dto.NewPeriodFormSignals(sharedmodels.FormTypeCreate, nil)
	key, err := getPeriodViewstoreKey(sharedmodels.FormTypeCreate, username, "")
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, key, signals)
}

// gets period data from the db, converts it to a form view, and saves it to the store
func refreshPeriodEditState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	periods *events.ReadModel,
	periodID string,
) error {
	model, err := periods.GetWithIDs(ctx, periodID)
	if err != nil {
		return err
	}
	signals := dto.NewPeriodFormSignals(sharedmodels.FormTypeEdit, model)
	key, err := getPeriodViewstoreKey(sharedmodels.FormTypeEdit, "", periodID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, key, signals)
}

func filterOutPeriod(periods []models.Period, excludeID string) []models.Period {
	filtered := make([]models.Period, 0, len(periods))
	for _, p := range periods {
		if p.ID != excludeID {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// buildScheduleViews constructs the schedule preview for a period.
// it includes a base view and one view per assigned student, with visibility from the schedules map
// if the map is empty, default visibility is true
func buildScheduleViews(
	ctx context.Context,
	l *slog.Logger,
	period models.Period,
	schedules map[string]bool,
	periodsReadModel *events.ReadModel,
	studentsReadModel *studentEvents.ReadModel,
) []scheduleDTO.PersonWithScheduleView {
	// base view (always visible, index 0)
	views := []scheduleDTO.PersonWithScheduleView{
		scheduleDTO.NewPersonScheduleView(
			"",
			sharedmodels.Person{GivenName: "base", FamilyName: "view"},
			[]models.Period{period},
			true,
			0,
		),
	}

	for idx, studentID := range period.StudentIDs {
		if studentID == "" {
			continue
		}
		student, err := studentsReadModel.GetByID(ctx, studentID)
		if err != nil {
			l.ErrorContext(ctx, "get student", "err", err)
			continue
		}
		visible, ok := schedules[student.Username]
		if !ok {
			visible = true
		}
		studentPeriods, err := periodsReadModel.ListPeriodsForStudent(ctx, studentID)
		if err != nil {
			l.ErrorContext(ctx, "list periods for student", "err", err)
			continue
		}
		studentPeriods = filterOutPeriod(studentPeriods, period.ID)
		views = append(views, scheduleDTO.NewPersonScheduleView(
			student.ID,
			student.Person,
			studentPeriods,
			visible,
			idx+1,
		))
	}
	return views
}

func createPeriodsListView(
	ctx context.Context,
	l *slog.Logger,
	periodReadModel events.ReadModel,
	educatorReadModel educatorEvents.ReadModel,
	studentReadModel studentEvents.ReadModel,
) shareddto.TableView {
	// get periods data from db
	periods, err := periodReadModel.ListWithIDs(ctx)
	if err != nil {
		l.ErrorContext(ctx, "create periods list view", "err", err)
		return shareddto.TableView{}
	}
	// create the views
	periodsWithData := make([]compositedto.PeriodWithData, len(periods))
	for i, period := range periods {

		periodsWithData[i] = compositedto.PeriodWithData{
			Period: period,
		}
		if len(period.EducatorIDs) > 0 {
			educatorViews := make([]educatorDTO.EducatorView, len(period.EducatorIDs))
			for i, educatorID := range period.EducatorIDs {
				educator, err := educatorReadModel.GetByID(ctx, educatorID)
				if err != nil {
					l.ErrorContext(ctx, "cplv get educator", "err", err, "eid", educatorID)
					return shareddto.TableView{}
				}
				educatorView := educatorDTO.NewView(educator)
				educatorViews[i] = educatorView
			}
			periodsWithData[i].Educators = educatorViews
		}
		if len(period.StudentIDs) > 0 {
			studentViews := make([]studentDTO.StudentView, len(period.StudentIDs))
			for i, studentID := range period.StudentIDs {
				student, err := studentReadModel.GetByID(ctx, studentID)
				if err != nil {
					l.ErrorContext(ctx, "cplv get student", "err", err, "sid", studentID)
					return shareddto.TableView{}
				}
				studentView := studentDTO.NewView(student)
				studentViews[i] = studentView
			}
			periodsWithData[i].Students = studentViews
		}
	}

	// create the table view
	return compositedto.NewPeriodWithDataTableView(periodsWithData)
}

func getPeriodViewstoreKey(
	formType sharedmodels.FormType,
	username string,
	periodID string,
) (string, error) {
	if formType == sharedmodels.FormTypeCreate && username != "" {
		return username + ".periods.create", nil
	} else if formType == sharedmodels.FormTypeEdit && periodID != "" {
		return periodID + ".edit", nil
	} else {
		return "", errors.New("error getting viewstore key")
	}
}
