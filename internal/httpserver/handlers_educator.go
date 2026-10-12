package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/events"
	"seek/internal/features/educators/models"
	"seek/internal/features/educators/pages"
	periodEvents "seek/internal/features/periods/events"
	scheduleDTO "seek/internal/features/schedules/dto"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	uifilters "seek/internal/ui/core/coreblocks/filters"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) educatorRoutes(r chi.Router) {
	// educators list
	r.Get("/educators", getEducatorsList(s.Logger))
	r.Get("/educators/stream", getEducatorsListStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Query("/educators", queryEducatorsList(s.Logger, s.ViewStore))
	// educator viewing
	r.Get("/educators/{username}", getEducatorView(s.Logger))
	r.Get("/educators/{username}/info", getEducatorViewInfo(s.Logger))
	r.Get("/educators/{username}/info/stream", getEducatorViewInfoStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Get("/educators/{username}/schedule", getEducatorViewSchedule(s.Logger))
	r.Get("/educators/{username}/schedule/stream", getEducatorViewScheduleStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Periods, s.ReadModels.Students))
	r.Get("/educators/{username}/caseload", getEducatorViewCaseload(s.Logger))
	r.Get("/educators/{username}/caseload/stream", getEducatorViewCaseloadStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	// educator creation
	r.Get("/educators/create", getEducatorCreate(s.Logger))
	r.Get("/educators/create/stream", getEducatorCreateStream(s.Logger, s.ViewStore))
	r.Query("/educators/create/validate", queryEducatorFormValidate(s.Logger, s.ViewStore))
	r.Post("/educators/create", postEducatorCreate(s.Logger, s.EventSaver))
	// educator editing
	r.Get("/educators/{username}/edit", getEducatorEdit(s.Logger))
	r.Get("/educators/{username}/edit/stream", getEducatorEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Query("/educators/{username}/edit/validate", queryEducatorFormValidate(s.Logger, s.ViewStore))
	r.Post("/educators/{username}/edit", postEducatorEdit(s.Logger, s.EventSaver, s.EventRetriever))
	// other educator stuff
	r.Delete("/educators/{id}", deleteEducator(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/e/{id}", getEducatorEditByID(s.Logger, s.ReadModels.Educators))
}

// GET request to /educators
func getEducatorsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewEducatorTableView([]models.Educator{})
		props := pages.ListPageProps{
			Signals: pages.ListPageSignals{},
			View:    view,
		}
		_ = pages.List(props).Render(ctx, w)
	}
}

// GET request to /educators/stream
func getEducatorsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		key := user.Username + ".educators.list"

		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "gels sub", "error", err)
			return
		}
		defer sub.Close()

		watcher, err := vs.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			l.ErrorContext(ctx, "gels watcher", "err", err)
			return
		}
		defer watcher.Stop()

		// seed only if empty
		signals, ok, err := viewstore.GetState[pages.ListPageSignals](ctx, vs, key)
		if err != nil {
			l.ErrorContext(ctx, "gels get state", "err", err, "key", key)
			return
		}
		if !ok {
			columns := make(map[string]bool, len(dto.EducatorTableColumns))
			for _, column := range dto.EducatorTableColumns {
				columns[column.Signal] = column.Visible
			}
			signals = pages.ListPageSignals{
				Table: pages.ListPageTableSignals{
					Columns: columns,
					Sort: shareddto.TableSort{
						Column:    "family_name",
						Direction: "ASC",
					},
					GroupedBy: "none",
				},
				Filters: *models.NewFilters(),
			}
			if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
				l.ErrorContext(ctx, "gels seed state", "err", err)
				return
			}
		}

		render := func(signals pages.ListPageSignals) {
			educators, err := listEducators(ctx, l, educatorReadModel, &signals.Table.Sort, signals.Table.GroupedBy, &signals.Filters)
			if err != nil {
				l.ErrorContext(ctx, "gels list", "err", err)
				return
			}
			props := pages.ListPageProps{
				Signals: signals,
				View:    dto.NewEducatorTableViewWithSort(educators, signals.Table.Columns, signals.Table.Sort, signals.Table.GroupedBy),
			}
			sse.PatchElementTempl(pages.List(props))
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				render(signals)
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				if err := entry.JSON(&signals); err != nil {
					l.ErrorContext(ctx, "gels watcher json", "err", err)
					return
				}
				render(signals)
			}
		}
	}
}

// QUERY request to /educators
// handles search and filter buttons
func queryEducatorsList(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.ListPageSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qel signals", "err", err)
			return
		}
		key := user.Username + ".educators.list"
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qel vs put state", "err", err)
		}
	}
}

// GET request to /students/{username}
// redirects to /students/{username}/info
func getEducatorView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		http.Redirect(w, r, fmt.Sprintf("/educators/%s/info", username), http.StatusFound)
	}
}

// GET request to /educators/{username}/info
func getEducatorViewInfo(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.View(dto.EducatorView{}, scheduleDTO.PersonWithScheduleView{}, []studentDTO.StudentView{}, "info").Render(ctx, w)
	}
}

// GET request to /educators/{username}/info/stream
func getEducatorViewInfoStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
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
			l.ErrorContext(ctx, "educator view stream subscribe", "err", err)
			return
		}
		defer sub.Close()
		educator, _ := educatorReadModel.GetByUsername(ctx, username)
		// watches the key value stream for ephemeral changes
		// lasts 5m
		if educator == nil {
			sse.PatchElementTempl(pages.NotFound())
			return
		}
		key := educator.ID + ".view"
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "educator view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshEducatorViewState(ctx, l, vs, username, educatorReadModel); err != nil {
			l.ErrorContext(ctx, "educator view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshEducatorViewState(ctx, l, vs, username, educatorReadModel); err != nil {
					if err.Error() == "educator not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "educator view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				educator := &models.Educator{}
				if err := entry.JSON(educator); err != nil {
					l.ErrorContext(ctx, "educator view stream json read in select", "err", err)
					return
				}
				view := dto.NewView(educator)
				sse.PatchElementTempl(pages.View(view, scheduleDTO.PersonWithScheduleView{}, []studentDTO.StudentView{}, "info"))
			}
		}
	}
}

// GET request to /educators/{username}/schedule
func getEducatorViewSchedule(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// create empty props for the page
		props := pages.EducatorViewSchedulePageProps{
			Educator: dto.NewView(&models.Educator{}),
			Periods:  []scheduleDTO.SchedulePeriodView{},
		}
		_ = pages.EducatorViewSchedule(props).Render(ctx, w)
	}
}

// GET request to /educators/{username}/caseload/stream
func getEducatorViewScheduleStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
	periodReadModel *periodEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
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
			l.ErrorContext(ctx, "educator view stream subscribe", "err", err)
			return
		}
		defer sub.Close()

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
			l.ErrorContext(ctx, "educator view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		educator, err := educatorReadModel.GetByUsername(ctx, username, events.WithRoles())
		if educator == nil {
			_ = pages.NotFound().Render(ctx, w)
			return
		}
		if err != nil {
			l.ErrorContext(ctx, "get educator view schedule db get", "error", err)
			return
		}

		// create the educator view and set the URL
		educatorView := dto.NewView(educator)

		// get periods and make views
		periods, err := periodReadModel.ListPeriodsForEducator(ctx, educator.ID)
		if err != nil {
			l.ErrorContext(ctx, "get educator view schedule db list periods", "err", err)
			return
		}
		periodViews := scheduleDTO.NewSchedulePeriodViews(periods...)
		for i, view := range periodViews {
			model, _ := periodReadModel.GetWithIDs(ctx, view.Period.ID)
			students, _ := studentReadModel.ListByIDs(ctx, model.StudentIDs)
			periodViews[i].Students = studentDTO.NewViews(students)
		}
		// create props for the page
		props := pages.EducatorViewSchedulePageProps{
			Educator: educatorView,
			Periods:  periodViews,
		}
		sse.PatchElementTempl(pages.EducatorViewSchedule(props))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshEducatorViewState(ctx, l, vs, username, educatorReadModel); err != nil {
					if err.Error() == "educator not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "educator view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				educatorView := &dto.EducatorView{}
				if err := entry.JSON(educatorView); err != nil {
					l.ErrorContext(ctx, "educator view stream json read in select", "err", err)
					return
				}
				// get periods and make views
				periods, err := periodReadModel.ListPeriodsForEducator(ctx, educator.ID)
				if err != nil {
					l.ErrorContext(ctx, "get educator view schedule db list periods", "err", err)
					return
				}
				periodViews := scheduleDTO.NewSchedulePeriodViews(periods...)
				for i, view := range periodViews {
					model, err := periodReadModel.GetWithIDs(ctx, view.Period.ID)
					if err != nil {
						l.ErrorContext(ctx, "gevss get period", "err", err, "pid", view.Period.ID)
					}
					students, err := studentReadModel.ListByIDs(ctx, model.StudentIDs)
					if err != nil {
						l.ErrorContext(ctx, "gevss get students", "err", err, "pid", len(model.StudentIDs))
					}
					periodViews[i].Students = studentDTO.NewViews(students)
				}
				// create props for the page
				props := pages.EducatorViewSchedulePageProps{
					Educator: *educatorView,
					Periods:  periodViews,
				}
				sse.PatchElementTempl(pages.EducatorViewSchedule(props))
			}
		}
	}
}

// GET request to /educators/{username}/caseload
func getEducatorViewCaseload(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.View(dto.EducatorView{}, scheduleDTO.PersonWithScheduleView{}, []studentDTO.StudentView{}, "caseload").Render(ctx, w)
	}
}

// GET request to /educators/{username}/caseload/stream
func getEducatorViewCaseloadStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
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
			l.ErrorContext(ctx, "educator view stream subscribe", "err", err)
			return
		}
		defer sub.Close()

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
			l.ErrorContext(ctx, "educator view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshEducatorViewState(ctx, l, vs, username, educatorReadModel); err != nil {
			l.ErrorContext(ctx, "educator view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshEducatorViewState(ctx, l, vs, username, educatorReadModel); err != nil {
					if err.Error() == "educator not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "educator view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				educator := &models.Educator{}
				if err := entry.JSON(educator); err != nil {
					l.ErrorContext(ctx, "educator view stream json read in select", "err", err)
					return
				}
				caseManager, err := educatorReadModel.GetByUsernameWithCaseload(ctx, educator.Username)
				if err != nil {
					l.ErrorContext(ctx, "educator view caseload stream", "err", err)
				}
				view := dto.NewView(educator)
				studentViews := studentDTO.NewViews(caseManager.Caseload)
				sse.PatchElementTempl(pages.View(view, scheduleDTO.PersonWithScheduleView{}, studentViews, "caseload"))
			}
		}
	}
}

// GET request to /educators/create
func getEducatorCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := pages.EducatorFormPageProps{
			FormType: sharedmodels.FormTypeCreate,
		}
		_ = pages.EducatorFormPage(view).Render(ctx, w)
	}
}

// GET request to /educators/create/stream
func getEducatorCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		key, err := getEducatorViewstoreKey(sharedmodels.FormTypeCreate, user.Username, "")
		if err != nil {
			l.ErrorContext(ctx, "gecs vs key", "err", err, "username", user.Username)
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
			l.ErrorContext(ctx, "gecs watch", "error", err)
			return
		}
		defer watcher.Stop()

		props, ok, err := viewstore.GetState[pages.EducatorFormPageProps](ctx, vs, key)
		if err != nil {
			l.ErrorContext(ctx, "gecs get state", "err", err, "key", key)
			return
		}
		if !ok {
			props := pages.EducatorFormPageProps{
				FormType: sharedmodels.FormTypeCreate,
			}
			if err := viewstore.PutState(ctx, vs, key, props); err != nil {
				l.ErrorContext(ctx, "gecs seed state", "err", err)
				return
			}
		}

		render := func(props pages.EducatorFormPageProps) {
			sse.PatchElementTempl(pages.EducatorFormPage(props))
		}

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					l.WarnContext(ctx, "gecs watcher updates channel closed")
					return
				}
				if err := entry.JSON(&props); err != nil {
					l.ErrorContext(ctx, "gecs watcher json", "error", err)
					return
				}
				render(props)
			}
		}
	}
}

// POST request to /educator/create
func postEducatorCreate(
	l *slog.Logger,
	saver eventstore.Saver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.EducatorFormPageProps{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "pec signals", "error", err)
			return
		}
		roleString := make([]string, len(signals.Educator.Roles))
		for i, role := range signals.Educator.Roles {
			roleString[i] = role.StringLong()
		}
		educator := events.EducatorState{
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Roles:      roleString,
		}
		cmd := events.CreateEducatorCommand{
			EducatorState: educator,
			Metadata:      eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.CreateEducatorCommandHandler(ctx, cmd, saver)
		if err != nil {
			l.ErrorContext(ctx, "pec command handler", "error", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/educators/%s", result.EventID))
	}
}

// GET request to /educators/{username}/edit
func getEducatorEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := pages.EducatorFormPageProps{
			FormType: sharedmodels.FormTypeEdit,
		}
		_ = pages.EducatorFormPage(view).Render(ctx, w)
	}
}

// GET request to /educator/{username}/stream
func getEducatorEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := chi.URLParam(r, "username")
		sse := newSSE(w, r)

		if username == "" {
			sse.Redirect("/404")
		}
		model, err := educatorReadModel.GetByUsername(ctx, username)

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(model.ID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "gees subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the educator edit view state kv
		key := model.ID + ".edit"
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "gees stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		props, ok, err := viewstore.GetState[pages.EducatorFormPageProps](ctx, vs, key)
		if err != nil {
			l.ErrorContext(ctx, "gees get state", "err", err, "key", key)
			return
		}
		if !ok {
			educator, err := getEducator(ctx, l, educatorReadModel, username)
			if err != nil {
				l.ErrorContext(ctx, "gees get educator", "err", err, "educator username", username)
				return
			}
			if educator == nil {
				sse.Redirect("/404")
			}
			props := pages.EducatorFormPageProps{
				FormType: sharedmodels.FormTypeEdit,
				Educator: *educator,
			}
			if err := viewstore.PutState(ctx, vs, key, props); err != nil {
				l.ErrorContext(ctx, "gees seed state", "err", err)
				return
			}
		}

		render := func(props pages.EducatorFormPageProps) {
			sse.PatchElementTempl(pages.EducatorFormPage(props))
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				render(props)
			case entry, ok := <-watcher.Updates():
				if !ok {
					l.WarnContext(ctx, "gees watcher updates channel closed")
					return
				}
				if err := entry.JSON(&props); err != nil {
					l.ErrorContext(ctx, "gees json read", "err", err)
					return
				}
				render(props)
			}
		}
	}
}

// POST request to /educators/{username}/edit
func postEducatorEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.EducatorFormPageProps{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post educator edit signal read", "err", err)
			return
		}
		roleString := make([]string, len(signals.Educator.Roles))
		for i, role := range signals.Educator.Roles {
			roleString[i] = role.StringLong()
		}
		educator := events.EducatorState{
			ID:         signals.Educator.ID,
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Username:   signals.Educator.Username,
			Roles:      roleString,
		}
		cmd := events.UpdateEducatorCommand{
			EducatorState: educator,
			Metadata:      eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateEducatorCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post educator edit command handler", "err", err)
			return
		}
		l.InfoContext(ctx, "post educator edit command handler", "eventID", result.EventID, "skipped", result.Skipped)
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/educators/%s", result.Educator.Username))
	}
}

// QUERY request to /educators/create/validate
// or /educators/edit/{id}
// TODO add actual validation
func queryEducatorFormValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.EducatorFormPageProps{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qefv signals", "err", err)
			return
		}
		key, err := getEducatorViewstoreKey(signals.FormType, user.Username, signals.Educator.ID)
		if err != nil {
			l.ErrorContext(ctx, "qefv key", "err", err, "form type", signals.FormType, "username", user.Username, "educator id", signals.Educator.ID)
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "educator create validate view store", "err", err)
		}
	}
}

// POST request to /educators/{username}
func deleteEducator(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		educatorID := chi.URLParam(r, "id")
		sse := newSSE(w, r)
		if educatorID == "" {
			l.ErrorContext(ctx, "de url param empty")
			sse.Redirect("/404")
			return
		}
		cmd := events.DeleteEducatorCommand{
			EducatorID: educatorID,
			Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.DeleteEducatorCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "de command handler", "err", err)
			return
		}
		l.InfoContext(ctx, "de educator deleted", "id", educatorID, "event", result.EventID)
		sse.Redirect("/educators")
	}
}

func getEducatorEditByID(
	l *slog.Logger,
	rm *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		model, err := rm.GetByID(ctx, id)
		if err != nil {
			l.ErrorContext(ctx, "geebid db", "err", err)
		}
		view := pages.EducatorFormPageProps{
			FormType: sharedmodels.FormTypeEdit,
			Educator: *model,
		}
		_ = pages.EducatorFormPage(view).Render(ctx, w)
	}
}

// helper functions

func refreshEducatorViewState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	username string,
	educatorReadModel *events.ReadModel,
) error {
	model, err := educatorReadModel.GetByUsername(ctx, username, events.WithRoles())
	if err != nil {
		return err
	}
	key := model.ID + ".view"
	return viewstore.PutState(ctx, vs, key, model)
}

func refreshEducatorEditState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	username string,
	educatorReadModel *events.ReadModel,
) error {
	model, err := educatorReadModel.GetByUsername(ctx, username, events.WithRoles())
	if err != nil {
		return err
	}
	view := pages.EducatorFormPageProps{
		FormType: sharedmodels.FormTypeEdit,
		Educator: *model,
	}
	key, err := getEducatorViewstoreKey(sharedmodels.FormTypeEdit, "", model.ID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, key, view)
}

func getEducator(
	ctx context.Context,
	_ *slog.Logger,
	rm *events.ReadModel,
	username string,
) (*models.Educator, error) {
	model, err := rm.GetByUsername(ctx, username, events.WithRoles())
	if err != nil {
		return nil, err
	}
	return model, nil
}

func listEducators(
	ctx context.Context,
	_ *slog.Logger,
	rm *events.ReadModel,
	sort *shareddto.TableSort,
	groupedBy string,
	filters *models.Filters,
) ([]models.Educator, error) {
	stringFilter := []string{}
	if filters != nil {
		stringFilter = models.ModelToString(filters.Roles)
	}
	if sort == nil {
		sort = &shareddto.TableSort{
			Column:    "family_name",
			Direction: "ASC",
		}
	}
	sortCol := sort.Column
	if sort.Column == "first_l" {
		sortCol = "given_name"
	}
	return rm.List(
		ctx,
		events.WithSortColumn(sortCol),
		events.WithSortDirection(sort.Direction),
		events.WithGroupedBy(groupedBy),
		events.WithSearchFilter(filters.Search),
		events.WithRoleFilter(stringFilter),
	)
}

func listEducatorsByIDs(
	ctx context.Context,
	l *slog.Logger,
	rm *events.ReadModel,
	ids []string,
) []models.Educator {
	educators, err := rm.ListByIDs(ctx, ids)
	if err != nil {
		l.ErrorContext(ctx, "list educators by ids", "err", err)
		return []models.Educator{}
	}
	return educators
}

func createEducatorSelectView(
	ctx context.Context,
	l *slog.Logger,
	educatorReadModel *events.ReadModel,
	filters *models.Filters,
	selected []string,
) dto.SelectView {
	educators, err := listEducators(ctx, l, educatorReadModel, nil, "", filters)
	if err != nil {
		l.ErrorContext(ctx, "cesv list educators", "err", err)
		return dto.SelectView{}
	}
	return dto.NewSelectView(nil, educators, selected...)
}

func roleFilterButtons(roleFilter []string) uifilters.FilterButtonsProps {
	buttons := make([]uifilters.FilterButtonProp, len(sharedmodels.EducatorRoleList))
	signalKey := "filter.role"
	for i, r := range sharedmodels.EducatorRoleList {
		active := slices.Contains(roleFilter, r.StringLong())
		buttons[i] = uifilters.FilterButtonProp{
			Label:       r.String(),
			SignalKey:   signalKey,
			SignalValue: r.StringLong(),
			Active:      active,
		}
	}
	return uifilters.FilterButtonsProps{
		Label:   "role",
		Buttons: buttons,
	}
}

func getEducatorViewstoreKey(
	formType sharedmodels.FormType,
	username string,
	educatorID string,
) (string, error) {
	if formType == sharedmodels.FormTypeCreate && username != "" {
		return username + ".educators.create", nil
	} else if formType == sharedmodels.FormTypeEdit && educatorID != "" {
		return educatorID + ".edit", nil
	} else {
		return "", errors.New("error getting viewstore key")
	}
}
