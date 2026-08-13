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
	scheduledto "seek/internal/features/schedules/dto"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) educatorRoutes(r chi.Router) {
	r.Get("/educators", getEducatorsList(s.Logger, s.ReadModels.Educators))
	r.Get("/educators/stream", s.getEducatorsListStream)
	r.Get("/educators/create", s.getEducatorCreate)
	r.Get("/educators/create/stream", s.getEducatorCreateStream)
	r.Post("/educators/create/validate", postEducatorCreateValidate(s.Logger, s.ViewStore))
	r.Post("/educators/create", s.postEducatorCreate)
	r.Get("/educators/{username}", getEducatorView())
	r.Get("/educators/{username}/info", getEducatorViewInfo(s.Logger))
	r.Get("/educators/{username}/info/stream", getEducatorViewInfoStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Educators))
	r.Get("/educators/{username}/schedule", s.getEducatorViewSchedule)
	r.Get("/educators/{username}/edit", s.getEducatorEdit)
	r.Get("/educators/{username}/edit/stream", s.getEducatorEditStream)
	r.Post("/educators/{username}/edit/validate", s.postEducatorEditValidate)
	r.Post("/educators/{username}/edit", s.postEducatorEdit)
	r.Delete("/educators/{username}", s.deleteEducator)
}

// GET request to /educators
func getEducatorsList(
	l *slog.Logger,
	rm *events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		educators, err := rm.ListWithRoles(ctx)
		if err != nil {
			l.ErrorContext(ctx, "educators list db list", "err", err)
			return
		}
		view := dto.NewEducatorTableView(educators)
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /educators/stream
func (s Server) getEducatorsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sse := newSSE(w, r)
	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any educators
	sub, err := s.Subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "educators list stream subscribe error", "error", err)
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
			educators, err := s.ReadModels.Educators.List(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			view := dto.NewEducatorTableView(educators)
			sse.PatchElementTempl(pages.List(view))
		}
	}
}

// GET request to /educators/create
func (s Server) getEducatorCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	empty := models.Educator{}
	view := dto.NewEducatorFormView(&empty)
	_ = pages.Create(view).Render(ctx, w)
}

// GET request to /educators/create/stream
func (s Server) getEducatorCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sse := newSSE(w, r)

	watcher, err := s.ViewStore.Watch(
		ctx,
		"neweducator",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "watcher error in educator create stream", "error", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates():
			if !ok {
				s.Logger.WarnContext(ctx, "educator watcher updates channel closed")
				return
			}
			model := &models.Educator{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "failed to unmarshal educator from view store", "error", err)
				return
			}
			view := dto.NewEducatorFormView(model)
			sse.PatchElementTempl(pages.Create(view))
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
		signals := &struct {
			Educator dto.EducatorFormView `json:"educator"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "educator create validate read signals", "err", err)
			return
		}
		l.Debug("pecv", "roles", signals.Educator.Role)
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
		if err := viewstore.PutState(ctx, vs, "neweducator", model); err != nil {
			l.ErrorContext(ctx, "educator create validate view store", "err", err)
		}

	}
}

// POST request to /educator/create
func (s Server) postEducatorCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Educator dto.EducatorFormView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post educator create signals", "error", err)
		return
	}
	result, err := events.CreateEducatorCommandHandler(ctx, events.CreateEducatorCommand{
		GivenName:  signals.Educator.GivenName,
		ChosenName: signals.Educator.ChosenName,
		FamilyName: signals.Educator.FamilyName,
		Email:      signals.Educator.Email,
		Roles:      strings.Split(signals.Educator.Role, ","),
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post educator create command handler", "error", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/educators/%s", result.EventID))
}

// GET request to /students/{username}
// redirects to /students/{username}/info
func getEducatorView() http.HandlerFunc {
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
		_ = pages.View(dto.EducatorView{}, scheduledto.PersonWithScheduleView{}, "info").Render(ctx, w)
	}
}

// GET request to /educators/{username}/info/stream
func getEducatorViewInfoStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel events.ReadModel,
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

		if err := refreshEducatorViewState(l, ctx, username, vs, educatorReadModel); err != nil {
			l.ErrorContext(ctx, "educator view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshEducatorViewState(l, ctx, username, vs, educatorReadModel); err != nil {
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
				view := dto.NewEducatorView(educator)
				sse.PatchElementTempl(pages.View(view, scheduledto.PersonWithScheduleView{}, "info"))
			}
		}
	}
}

// GET request to /educators/{username}/schedule
func (s Server) getEducatorViewSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get the username from the URL and get the educator from the db
	username := chi.URLParam(r, "username")
	educator, err := s.ReadModels.Educators.GetByUsername(ctx, username)
	if educator == nil {
		_ = pages.NotFound().Render(ctx, w)
		return
	}
	if err != nil {
		s.Logger.ErrorContext(ctx, "get educator view schedule db get", "error", err)
		return
	}

	// create the educator view and set the URL
	educatorView := dto.NewEducatorView(educator)

	// get periods for the educator and make views
	periods, err := s.ReadModels.Periods.ListPeriodsForEducator(ctx, educator.ID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get educator view schedule db list periods", "err", err)
		return
	}

	scheduleView := scheduledto.NewPersonScheduleView(educator.Person, periods, true, 1)
	_ = pages.View(educatorView, scheduleView, "schedule").Render(ctx, w)
}

// GET request to /educators/{username}/edit
func (s Server) getEducatorEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	educator, err := s.ReadModels.Educators.GetByUsername(ctx, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if educator == nil {
		_ = pages.NotFound().Render(ctx, w)
		return
	}

	view := dto.NewEducatorFormView(educator)
	_ = pages.Edit(view).Render(ctx, w)
}

// GET request to /educator/{username}/stream
func (s Server) getEducatorEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(username), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "educator edit stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	// watches the educator edit view state kv
	watcher, err := s.ViewStore.Watch(
		ctx,
		username+".edit",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "educator edit stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshEducatorEditState(ctx, username); err != nil {
				if err.Error() == "educator not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				s.Logger.ErrorContext(ctx, "educator edit stream refresh", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			model := &models.Educator{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "educator edit stream json read", "err", err)
				return
			}
			view := dto.NewEducatorFormView(model)
			sse.PatchElementTempl(pages.Edit(view))
		}
	}
}

// POST request to /educators/{username}/edit/validate
func (s Server) postEducatorEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := chi.URLParam(r, "username")
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
	viewstore.PutState(ctx, s.ViewStore, username, model)
}

// POST request to /educators/{username}/edit
func (s Server) postEducatorEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	username := chi.URLParam(r, "username")
	signals := &struct {
		Educator dto.EducatorFormView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post educator edit signal read", "err", err)
		return
	}
	result, err := events.UpdateEducatorCommandHandler(ctx, events.UpdateEducatorCommand{
		EducatorID: signals.Educator.ID,
		GivenName:  signals.Educator.GivenName,
		ChosenName: signals.Educator.ChosenName,
		FamilyName: signals.Educator.FamilyName,
		Roles:      strings.Split(signals.Educator.Role, ","),
		Email:      signals.Educator.Email,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post educator edit command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.InfoContext(ctx, "post educator edit command handler", "skipped", result.Skipped)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/educators/%s", username))
}

// POST request to /educators/{username}
func (s Server) deleteEducator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	username := chi.URLParam(r, "username")
	educator, err := s.ReadModels.Educators.GetByUsername(ctx, username)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete educator db get by username", "err", err)
		return
	}
	result, err := events.DeleteEducatorCommandHandler(ctx, events.DeleteEducatorCommand{
		EducatorID: educator.ID,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete educator command handler", "err", err)
		return
	}
	s.Logger.InfoContext(ctx, "delete educator student deleted", "id", educator.ID, "event", result.EventID)
	sse := newSSE(w, r)
	sse.Redirect("/educators")
}

// helper functions

func (s Server) refreshEducatorViewState(ctx context.Context, username string) error {
	educator, err := s.ReadModels.Educators.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, educator.ID+".view", educator)
}

func refreshEducatorViewState(
	_ *slog.Logger,
	ctx context.Context,
	username string,
	vs viewstore.Store,
	educatorReadModel events.ReadModel,
) error {
	educator, err := educatorReadModel.GetByUsernameWithRoles(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, educator.Username+".view", educator)
}

func (s Server) refreshEducatorEditState(ctx context.Context, username string) error {
	educator, err := s.ReadModels.Educators.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, educator.ID+".edit", educator)
}
