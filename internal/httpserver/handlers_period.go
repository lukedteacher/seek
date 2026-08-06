package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	eevents "seek/internal/features/educators/events"
	epevents "seek/internal/features/educators_periods/events"
	"seek/internal/features/periods/dto"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	"seek/internal/features/periods/pages"
	scheduleDTO "seek/internal/features/schedules/dto"
	sdto "seek/internal/features/students/dto"
	sevents "seek/internal/features/students/events"
	spevents "seek/internal/features/students_periods/events"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get(
		"/periods",
		getPeriodsList(s.Logger),
	)
	r.Get(
		"/periods/stream",
		getPeriodsListStream(
			s.Logger,
			s.ReadModels.Periods,
			s.ReadModels.EducatorPeriods,
			s.ReadModels.StudentPeriods,
			s.Subscriber,
		),
	)
	r.Get(
		"/periods/create",
		getPeriodCreate(s.Logger),
	)
	r.Get(
		"/periods/create/stream",
		getPeriodCreateStream(
			s.Logger,
			s.ReadModels.Periods,
			s.ReadModels.StudentPeriods,
			s.ReadModels.Students,
			s.ReadModels.Educators,
			s.ViewStore,
		),
	)
	r.Post(
		"/periods/create/validate",
		postPeriodCreateValidate(
			s.Logger,
			s.ViewStore,
		),
	)
	r.Post(
		"/periods/create/validate/{field}",
		postPeriodCreateValidateField(
			s.Logger,
			s.ViewStore,
		),
	)
	r.Post(
		"/periods/create",
		postPeriodCreate(
			s.Logger,
			s.EventSaver,
			s.EventRetriever,
		),
	)
	r.Get(
		"/periods/{id}",
		getPeriodView(s.Logger),
	)
	r.Get(
		"/periods/{id}/stream",
		getPeriodViewStream(
			s.Logger,
			s.Subscriber,
			s.ViewStore,
			s.ReadModels.Periods,
			s.ReadModels.Educators,
			s.ReadModels.Students,
		),
	)
	r.Get(
		"/periods/{id}/edit",
		getPeriodEdit(s.Logger),
	)
	r.Get(
		"/periods/{id}/edit/stream",
		getPeriodEditStream(
			s.Logger,
			s.ReadModels.Periods,
			s.ReadModels.Students,
			s.ReadModels.Educators,
			s.Subscriber,
			s.ViewStore,
		),
	)
	r.Post(
		"/periods/{id}/edit/validate",
		postPeriodEditValidate(s.Logger, s.ViewStore),
	)
	r.Post(
		"/periods/{id}/edit",
		postPeriodEdit(s.Logger, s.EventSaver, s.EventRetriever),
	)
	r.Post(
		"/periods/{id}/archive",
		postPeriodArchive(s.Logger, s.EventSaver, s.EventRetriever),
	)
	r.Delete(
		"/periods/{id}",
		deletePeriod(s.Logger, s.EventSaver, s.EventRetriever),
	)
}

// GET request to /periods
// renders an empty template
// SSE will populate data
func getPeriodsList(
	l *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewPeriodTableView([]models.Period{})
		view.URL = "/periods"
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /periods/stream
// populates data
func getPeriodsListStream(
	l *slog.Logger,
	periodReadModel *events.ReadModel,
	educatorPeriodReadModel *epevents.ReadModel,
	studentPeriodReadModel *spevents.ReadModel,
	sub MessageSubscriber,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to any periods
		sub, err := sub.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "periods list stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		buildView := func() (shareddto.TableView, error) {
			periods, err := periodReadModel.List(ctx)
			if err != nil {
				return shareddto.TableView{}, err
			}
			for i := range periods {
				ids, err := educatorPeriodReadModel.ListEducatorIDsForPeriod(ctx, periods[i].ID)
				if err != nil {
					return shareddto.TableView{}, err
				}
				periods[i].EducatorIDs = ids
				ids, err = studentPeriodReadModel.ListStudentIDsForPeriod(ctx, periods[i].ID)
				if err != nil {
					return shareddto.TableView{}, err
				}
				periods[i].StudentIDs = ids
			}
			view := dto.NewPeriodTableView(periods)
			view.URL = "/periods"
			return view, nil
		}
		view, err := buildView()
		if err != nil {
			l.ErrorContext(ctx, "build initial view", "err", err)
			return
		}
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				// for now just reloads the page
				// consider adding a view store for the list

				view, err := buildView()
				if err != nil {
					l.ErrorContext(ctx, "build updated view", "err", err)
					continue
				}
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /periods/create
func getPeriodCreate(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.PeriodFormView{URL: "/periods/create"}
		_ = pages.Create(view, nil).Render(ctx, w)
	}
}

// GET request to /periods/create/stream
func getPeriodCreateStream(
	l *slog.Logger,
	prm *events.ReadModel,
	sprm *spevents.ReadModel,
	srm *sevents.ReadModel,
	erm *eevents.ReadModel,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// initial load: empty period, empty schedules
		empty, err := models.NewPeriod()
		if err != nil {
			l.ErrorContext(ctx, "new period", "err", err)
			return
		}

		scheduleViews := buildScheduleViews(ctx, *empty, nil, prm, srm, l)
		view := buildPeriodFormView(ctx, empty, erm, srm, l)
		view.URL = "/periods/create"
		sse.PatchElementTempl(pages.Create(view, scheduleViews))

		// watch for view store changes
		watcher, err := vs.Watch(ctx, "new", viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			l.ErrorContext(ctx, "watch", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &struct {
					Period    dto.PeriodFormView `json:"period"`
					Schedules map[string]bool    `json:"schedules"`
				}{}
				if err := entry.JSON(signals); err != nil {
					l.Error("json decode", "err", err)
					return
				}
				period := signals.Period.ToPeriod()
				scheduleViews := buildScheduleViews(ctx, period, signals.Schedules, prm, srm, l)
				view := buildPeriodFormView(ctx, &period, erm, srm, l)
				view.URL = "/periods/create"
				sse.PatchElementTempl(pages.Create(view, scheduleViews))
			}
		}
	}
}

// POST request to /periods/create/validate
// reads datastar signals from the form and saves them to the view store.
// this allows the SSE stream to detect changes and refresh the form preview.
func postPeriodCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		signals := &struct {
			Period    dto.PeriodFormView `json:"period"`
			Schedules map[string]bool    `json:"schedules"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "period create validate signals", "err", err.Error())
			return
		}

		// store the signals under key "new" so the create stream can react
		if err := viewstore.PutState(ctx, vs, "new", signals); err != nil {
			l.ErrorContext(ctx, "post period create validate viewstore", "err", err)
		}
	}
}

func postPeriodCreateValidateField(l *slog.Logger, vs viewstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			Period dto.PeriodFormView `json:"period"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "pcvf signal read", "error", err)
			return
		}
		field := chi.URLParam(r, "field")
		l.DebugContext(ctx, "validating field", "field", field, "st", signals.Period.ServiceType)

		// convert the form view to a temporary Period for business logic
		temp := signals.Period.ToPeriod()

		switch field {
		case "starttime":
			// update start time → recalculate end time using current duration
			temp.UpdateStartTime(signals.Period.StartTime)
			// copy back the recalculated end time (duration unchanged)
			signals.Period.EndTime = temp.EndTime
			l.DebugContext(ctx, "start time updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

		case "endtime":
			// update end time → recalculate duration using current start time
			temp.UpdateEndTime(signals.Period.EndTime)
			// copy back the new duration (end time already set)
			signals.Period.Duration = temp.Duration
			signals.Period.StartTime = temp.StartTime
			l.DebugContext(ctx, "end time updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

		case "duration":
			// update duration → recalculate end time using current start time
			temp.UpdateDuration(signals.Period.Duration)
			// copy back the recalculated end time
			signals.Period.EndTime = temp.EndTime
			l.DebugContext(ctx, "duration updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

		default:
			l.WarnContext(ctx, "unknown validation field", "field", field)
			http.Error(w, "unknown field", http.StatusBadRequest)
			return
		}

		// save the updated signals to the view store so the SSE can refresh the form
		if err := viewstore.PutState(ctx, vs, "new", signals); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
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

		// sync educators (proposed list from form)
		secmd := epevents.SyncEducatorsInPeriodCommand{
			PeriodID:            periodID,
			ProposedEducatorIDs: strings.Split(signals.Period.EducatorIDs, ","),
		}
		if _, err := epevents.SyncEducatorsInPeriodCommandHandler(ctx, secmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period create sync educators", "err", err)
		}

		// sync students
		spcmd := spevents.SyncStudentsInPeriodCommand{
			PeriodID:           periodID,
			ProposedStudentIDs: strings.Split(signals.Period.StudentIDs, ","),
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
func getPeriodView(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		periodID := chi.URLParam(r, "id")

		// minimal empty view with URL for SSE
		view := dto.PeriodView{URL: fmt.Sprintf("/periods/%s", periodID)}
		_ = pages.View(view).Render(ctx, w)
	}
}
func getPeriodViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	periodReadModel *events.ReadModel,
	educatorReadModel *eevents.ReadModel,
	studentReadModel *sevents.ReadModel,
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

		if err := refreshPeriodViewState(ctx, l, periodID, "view", periodReadModel, educatorReadModel, studentReadModel, vs); err != nil {
			l.ErrorContext(ctx, "get period view stream refresh", "err", err)
			return
		}

		if err != nil {
			l.ErrorContext(ctx, "get period view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshPeriodViewState(ctx, l, periodID, "view", periodReadModel, educatorReadModel, studentReadModel, vs); err != nil {
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
				view.URL = fmt.Sprintf("/periods/%s", periodID)
				for i := range period.EducatorIDs {
					educator, _ := educatorReadModel.GetByID(ctx, period.EducatorIDs[i])
					educatorView := edto.NewEducatorView(educator)
					view.Educators = append(view.Educators, educatorView)
				}
				for i := range period.StudentIDs {
					student, _ := studentReadModel.GetByID(ctx, period.StudentIDs[i])
					studentView := sdto.NewStudentView(student)
					view.Students = append(view.Students, studentView)
				}
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

func getPeriodEdit(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		periodID := chi.URLParam(r, "id")
		// creates a minimal view and sets the URL
		view := dto.PeriodFormView{URL: fmt.Sprintf("/periods/%s/edit", periodID)}
		_ = pages.Edit(view, nil).Render(ctx, w)
	}
}

// GET request to /periods/{id}/edit/stream
// establishes an SSE connection that updates the edit form when the period or view state changes.
func getPeriodEditStream(
	l *slog.Logger,
	periodsReadModel *events.ReadModel,
	studentsReadModel *sevents.ReadModel,
	educatorsReadModel *eevents.ReadModel,
	subscriber MessageSubscriber,
	viewStore viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		periodID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// --- subscribe to event store changes ---
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "period edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// --- watch the view store for form validations ---
		watcher, err := viewStore.Watch(ctx, periodID+".edit", viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			l.ErrorContext(ctx, "edit view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		period, _ := periodsReadModel.GetWithIDs(ctx, periodID)
		scheduleViews := buildScheduleViews(
			ctx,
			*period,
			map[string]bool{},
			periodsReadModel,
			studentsReadModel,
			l,
		)
		view := buildPeriodFormView(
			ctx,
			period,
			educatorsReadModel,
			studentsReadModel,
			l,
		)
		view.URL = fmt.Sprintf("/periods/%s/edit", periodID)
		sse.PatchElementTempl(pages.Edit(view, scheduleViews))

		for {
			select {
			case <-ctx.Done():
				return

			case <-notifier.Signal():
				// period changed via event – refresh the view state and re‑render
				if err := refreshPeriodViewState(ctx, l, periodID, "edit", periodsReadModel, educatorsReadModel, studentsReadModel, viewStore); err != nil {
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
				signals := &struct {
					Period    dto.PeriodFormView `json:"period"`
					Schedules map[string]bool    `json:"schedules"`
				}{}
				if err := entry.JSON(signals); err != nil {
					l.Error("period edit stream json", "err", err)
					return
				}

				period := signals.Period.ToPeriod()
				scheduleViews := buildScheduleViews(
					ctx,
					period,
					signals.Schedules,
					periodsReadModel,
					studentsReadModel,
					l,
				)
				view := buildPeriodFormView(
					ctx,
					&period,
					educatorsReadModel,
					studentsReadModel,
					l,
				)
				view.URL = fmt.Sprintf("/periods/%s/edit", periodID)
				sse.PatchElementTempl(pages.Edit(view, scheduleViews))
			}
		}
	}
}

// POST request to /periods/{id}/edit/validate
// reads datastar signals from the form and saves them to the view store.
// this allows the SSE stream to detect changes and refresh the edit form preview.
func postPeriodEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		periodID := chi.URLParam(r, "id")

		signals := &struct {
			FormView  dto.PeriodFormView `json:"period"`
			Schedules map[string]bool    `json:"schedules"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "period edit validate signals", "err", err)
			return
		}

		// store the signals under a key scoped to the period
		key := periodID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "post period edit validate viewstore", "err", err)
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

		signals := &struct {
			Period dto.PeriodFormView `json:"period"`
		}{}
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
			DaysBitmask: signals.Period.Days.ToBitmask(),
			Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdatePeriodCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post period edit command handler", "err", err)
			return
		}
		if result.Skipped {
			l.InfoContext(ctx, "post period edit command handler", "skipped", result.Skipped)
		}

		// sync educators
		secmd := epevents.SyncEducatorsInPeriodCommand{
			PeriodID:            periodID,
			ProposedEducatorIDs: strings.Split(signals.Period.EducatorIDs, ","),
		}
		if _, err := epevents.SyncEducatorsInPeriodCommandHandler(ctx, secmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post period edit sync educators", "err", err)
		}

		// sync students
		spcmd := spevents.SyncStudentsInPeriodCommand{
			PeriodID:           periodID,
			ProposedStudentIDs: strings.Split(signals.Period.StudentIDs, ","),
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
	l *slog.Logger,
	periodID string,
	keySuffix string,
	periods *events.ReadModel,
	educators *eevents.ReadModel,
	students *sevents.ReadModel,
	vs viewstore.Store,
) error {
	period, err := periods.GetWithIDs(ctx, periodID)
	if err != nil {
		return err
	}
	buildPeriodFormView(ctx, period, educators, students, l)
	return viewstore.PutState(ctx, vs, period.ID+"."+keySuffix, period)
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

// buildPeriodFormView fetches all educators and students and constructs a PeriodFormView.
func buildPeriodFormView(
	ctx context.Context,
	period *models.Period,
	educatorsReadModel *eevents.ReadModel,
	studentsReadModel *sevents.ReadModel,
	logger *slog.Logger,
) dto.PeriodFormView {
	educators, err := educatorsReadModel.List(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "list educators", "err", err)
	}
	students, err := studentsReadModel.List(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "list students", "err", err)
	}
	return dto.NewPeriodFormView(period, students, educators)
}

// buildScheduleViews constructs the schedule preview for a period.
// it includes a base view and one view per assigned student, with visibility from the schedules map.
func buildScheduleViews(
	ctx context.Context,
	period models.Period,
	schedules map[string]bool,
	periodsReadModel *events.ReadModel,
	studentsReadModel *sevents.ReadModel,
	logger *slog.Logger,
) []scheduleDTO.PersonWithScheduleView {
	// base view (always visible, index 0)
	views := []scheduleDTO.PersonWithScheduleView{
		scheduleDTO.NewPersonScheduleView(
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
			logger.ErrorContext(ctx, "get student", "err", err)
			continue
		}
		visible, ok := schedules[student.Username]
		if !ok {
			logger.Debug("testing", "ok", ok)
			visible = true
		}
		studentPeriods, err := periodsReadModel.ListPeriodsForStudent(ctx, studentID)
		if err != nil {
			logger.ErrorContext(ctx, "list periods for student", "err", err)
			continue
		}
		studentPeriods = filterOutPeriod(studentPeriods, period.ID)
		views = append(views, scheduleDTO.NewPersonScheduleView(
			student.Person,
			studentPeriods,
			visible,
			idx+1,
		))
	}
	return views
}
