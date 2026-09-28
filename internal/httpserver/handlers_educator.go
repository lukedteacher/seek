package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/events"
	"seek/internal/features/educators/models"
	"seek/internal/features/educators/pages"
	periodEvents "seek/internal/features/periods/events"
	scheduleDTO "seek/internal/features/schedules/dto"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) educatorRoutes(r chi.Router) {
	r.Get("/educators", getEducatorsList(s.Logger))
	r.Get("/educators/stream", getEducatorsListStream(s.Logger, s.Subscriber, s.ReadModels.Educators))
	r.Get("/educators/create", getEducatorCreate(s.Logger))
	r.Get("/educators/create/stream", getEducatorCreateStream(s.Logger, s.ViewStore))
	r.Post("/educators/create/validate", postEducatorCreateValidate(s.Logger, s.ViewStore))
	r.Post("/educators/create", postEducatorCreate(s.Logger, s.EventSaver))
	r.Get("/educators/{username}", getEducatorView(s.Logger))
	r.Get("/educators/{username}/info", getEducatorViewInfo(s.Logger))
	r.Get("/educators/{username}/info/stream", getEducatorViewInfoStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Get("/educators/{username}/schedule", getEducatorViewSchedule(s.Logger))
	r.Get("/educators/{username}/schedule/stream", getEducatorViewScheduleStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Periods, s.ReadModels.Students))
	r.Get("/educators/{username}/caseload", getEducatorViewCaseload(s.Logger))
	r.Get("/educators/{username}/caseload/stream", getEducatorViewCaseloadStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Get("/educators/{username}/edit", getEducatorEdit(s.Logger))
	r.Get("/educators/{username}/edit/stream", getEducatorEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Post("/educators/{username}/edit/validate", postEducatorEditValidate(s.Logger, s.ViewStore))
	r.Post("/educators/{username}/edit", postEducatorEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Delete("/educators/{id}", deleteEducator(s.Logger, s.EventSaver, s.EventRetriever, s.ReadModels.Educators))
	r.Get("/e/{id}", getEducatorEditByID(s.Logger, s.ReadModels.Educators))
}

// GET request to /educators
func getEducatorsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewEducatorTableView([]models.Educator{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /educators/stream
func getEducatorsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)
		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to any educators
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "educators list stream subscribe error", "error", err)
			return
		}
		defer sub.Close()

		educators, err := educatorReadModel.ListWithRoles(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		view := dto.NewEducatorTableView(educators)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				// for now just reloads the page
				// consider adding a view store for the list
				educators, err := educatorReadModel.ListWithRoles(ctx)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				view := dto.NewEducatorTableView(educators)
				sse.PatchElementTempl(pages.List(view))
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
		view := dto.NewEducatorFormView(&models.Educator{})
		_ = pages.Create(view).Render(ctx, w)
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

		key := user.Username + ".educators.create"
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "watcher error in educator create stream", "error", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					l.WarnContext(ctx, "educator watcher updates channel closed")
					return
				}
				model := &models.Educator{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "failed to unmarshal educator from view store", "error", err)
					return
				}
				view := dto.NewEducatorFormView(model)
				sse.PatchElementTempl(pages.Create(view))
			}
		}
	}
}

// POST request to /educators/create/validate
// validates the current state of the educator form via signals
// saves the state to a view store for SSE updates
func postEducatorCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			Educator dto.EducatorFormView `json:"educator"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "educator create validate read signals", "err", err)
			return
		}
		rolesStrings := strings.Split(signals.Educator.Role, ",")
		roles := make([]sharedmodels.EducatorRole, len(rolesStrings))
		for i, role := range rolesStrings {
			roles[i] = sharedmodels.EducatorRole(role)
		}
		model := models.Educator{
			ID: signals.Educator.ID,
			Person: sharedmodels.Person{
				GivenName:  signals.Educator.GivenName,
				ChosenName: signals.Educator.ChosenName,
				FamilyName: signals.Educator.FamilyName,
				Email:      signals.Educator.Email,
			},
			Roles: roles,
		}
		// saves the state to a view store so that the SSE can update
		// TODO look into a better name for the channel
		key := user.Username + ".educators.create"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "educator create validate view store", "err", err)
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
		signals := &struct {
			Educator dto.EducatorFormView `json:"educator"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post educator create signals", "error", err)
			return
		}
		educator := events.EducatorState{
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Roles:      strings.Split(signals.Educator.Role, ","),
		}
		cmd := events.CreateEducatorCommand{
			EducatorState: educator,
			Metadata:      eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.CreateEducatorCommandHandler(ctx, cmd, saver)
		if err != nil {
			l.ErrorContext(ctx, "post educator create command handler", "error", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/educators/%s", result.EventID))
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

// GET request to /educators/{username}/edit
func getEducatorEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewEducatorFormView(&models.Educator{})
		_ = pages.Edit(view).Render(ctx, w)
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
			l.ErrorContext(ctx, "educator edit stream subscribe", "err", err)
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
			l.ErrorContext(ctx, "educator edit stream watcher", "err", err)
			return
		}
		defer watcher.Stop()
		if err := refreshEducatorEditState(ctx, l, vs, username, educatorReadModel); err != nil {
			if err.Error() == "educator not found" {
				sse.PatchElementTempl(pages.NotFound())
			}
			l.ErrorContext(ctx, "educator edit stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				if err := refreshEducatorEditState(ctx, l, vs, username, educatorReadModel); err != nil {
					if err.Error() == "educator not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "educator edit stream refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.Educator{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "educator edit stream json read", "err", err)
					return
				}
				view := dto.NewEducatorFormView(model)
				sse.PatchElementTempl(pages.Edit(view))
			}
		}
	}
}

// POST request to /educators/{username}/edit/validate
func postEducatorEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			Educator dto.EducatorFormView `json:"educator"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rolesStringSlice := strings.Split(signals.Educator.Role, ",")
		roles := make([]sharedmodels.EducatorRole, len(rolesStringSlice))
		for i, role := range rolesStringSlice {
			roles[i] = sharedmodels.EducatorRole(role)
		}
		model := models.Educator{
			ID:     signals.Educator.ID,
			Person: signals.Educator.Person,
			Roles:  roles,
		}
		key := model.ID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "post homeroom create validate viewstore", "err", err)
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
		signals := &struct {
			Educator dto.EducatorFormView `json:"educator"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post educator edit signal read", "err", err)
			return
		}
		educator := events.EducatorState{
			ID:         signals.Educator.ID,
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Username:   signals.Educator.Username,
			Roles:      strings.Split(signals.Educator.Role, ","),
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

// POST request to /educators/{username}
func deleteEducator(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	educatorReadModel *events.ReadModel,
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
		view := dto.NewEducatorFormView(model)
		_ = pages.Edit(view).Render(ctx, w)
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
	key := model.ID + ".edit"
	return viewstore.PutState(ctx, vs, key, model)
}

func listEducators(
	ctx context.Context,
	l *slog.Logger,
	rm *events.ReadModel,
	filter *dto.Filter,
) ([]models.Educator, error) {
	if filter != nil {
		return rm.List(ctx, events.WithSearchFilter(filter.Search))
	}
	return rm.List(ctx)
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
	filter *dto.Filter,
	selected []string,
) dto.SelectView {
	educators, err := listEducators(ctx, l, educatorReadModel, filter)
	if err != nil {
		l.ErrorContext(ctx, "cesv list educators", "err", err)
		return dto.SelectView{}
	}
	return dto.NewSelectView(nil, educators, selected)
}
