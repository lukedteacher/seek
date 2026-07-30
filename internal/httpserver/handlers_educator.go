package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/events"
	"seek/internal/features/educators/models"
	"seek/internal/features/educators/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) educatorRoutes(r chi.Router) {
	r.Get("/educators", s.getEducatorsList)
	r.Get("/educators/stream", s.getEducatorsListStream)
	r.Get("/educators/create", s.getEducatorCreate)
	r.Get("/educators/create/stream", s.getEducatorCreateStream)
	r.Post("/educators/create/validate", s.postEducatorCreateValidate)
	r.Post("/educators/create", s.postEducatorCreate)
	r.Get("/educators/{id}", s.getEducatorView)
	r.Get("/educators/{id}/stream", s.getEducatorViewStream)
	r.Get("/educators/{id}/edit", s.getEducatorEdit)
	r.Get("/educators/{id}/edit/stream", s.getEducatorEditStream)
	r.Post("/educators/{id}/edit/validate", s.postEducatorEditValidate)
	r.Post("/educators/{id}/edit", s.postEducatorEdit)
	r.Delete("/educators/{id}", s.deleteEducator)
}

// GET request to /educators
func (s Server) getEducatorsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educators, err := s.Educators.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "educators list db list", "err", err)
		return
	}
	view := dto.NewEducatorTableView(educators)
	view.URL = "/educators"
	_ = pages.List(user, view).Render(ctx, w)
}

// GET request to /educators/stream
func (s Server) getEducatorsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
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
			educators, err := s.Educators.List(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			view := dto.NewEducatorTableView(educators)
			view.URL = "/educators"
			sse.PatchElementTempl(pages.List(user, view))
		}
	}
}

// GET request to /educators/create
func (s Server) getEducatorCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.Educator{}
	view := dto.NewEducatorView(&empty)
	view.URL = "/educators/create"
	_ = pages.Create(user, *view).Render(ctx, w)
}

// GET request to /educators/create/stream
func (s Server) getEducatorCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
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
			view := dto.NewEducatorView(model)
			view.URL = "/educators/create"
			sse.PatchElementTempl(pages.Create(user, *view))
		}
	}
}

// POST request to /educators/create/validate
func (s Server) postEducatorCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "educator create validate read signals", "err", err)
		return
	}
	model := models.Educator{
		ID: signals.Educator.ID,
		Person: sharedmodels.Person{
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
		},
		Role: signals.Educator.Role,
	}
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "neweducator", model); err != nil {
		s.Logger.ErrorContext(ctx, "educator create validate view store", "err", err)
	}
}

// POST request to /educator/create
func (s Server) postEducatorCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
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
		Role:       signals.Educator.Role,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post educator create command handler", "error", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/educators/%s", result.EventID))
}

// GET request to /educators/{id}
func (s Server) getEducatorView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	educator, err := s.Educators.Get(ctx, educatorID)
	if educator == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}
	if err != nil {
		s.Logger.ErrorContext(ctx, "get educator view", "error", err)
		return
	}
	view := dto.NewEducatorView(educator)
	view.URL = fmt.Sprintf("/educators/%s", educatorID)
	_ = pages.View(user, *view).Render(ctx, w)
}

// GET request to /educators/{id}/stream
func (s Server) getEducatorViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(educatorID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "educator view stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshEducatorViewState(ctx, educatorID); err != nil {
		s.Logger.ErrorContext(ctx, "educator view stream refresh", "err", err)
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		educatorID+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "educator view stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshEducatorViewState(ctx, educatorID); err != nil {
				if err.Error() == "educator not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				s.Logger.ErrorContext(ctx, "educator view stream refresh in select", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.Educator{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "educator view stream json read in select", "err", err)
				return
			}
			view := dto.NewEducatorView(model)
			view.URL = fmt.Sprintf("/educators/%s", educatorID)
			sse.PatchElementTempl(pages.View(user, *view))
		}
	}
}

// GET request to /educators/{id}/edit
func (s Server) getEducatorEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	model, err := s.Educators.Get(ctx, educatorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}

	view := dto.NewEducatorView(model)
	view.URL = fmt.Sprintf("/educators/%s/edit", educatorID)
	_ = pages.Edit(user, *view).Render(ctx, w)
}

// GET request to /educator/{id}/stream
func (s Server) getEducatorEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(educatorID), func(context.Context, []byte) {
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
		educatorID+".edit",
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
			if err := s.refreshEducatorEditState(ctx, educatorID); err != nil {
				if err.Error() == "educator not found" {
					sse.PatchElementTempl(pages.NotFound(user))
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
			view := dto.NewEducatorView(model)
			view.URL = fmt.Sprintf("/educators/%s/edit", educatorID)
			sse.PatchElementTempl(pages.Edit(user, *view))
		}
	}
}

// POST request to /educators/{id}/edit/validate
func (s Server) postEducatorEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	educatorID := chi.URLParam(r, "id")
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.Educator{
		ID:     signals.Educator.ID,
		Person: signals.Educator.Person,
		Role:   signals.Educator.Role,
	}
	viewstore.PutState(ctx, s.ViewStore, educatorID, model)
}

// POST request to /educators/{id}/edit
func (s Server) postEducatorEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post educator edit signal read", "err", err)
		return
	}
	educatorID := chi.URLParam(r, "id")
	result, err := events.UpdateEducatorCommandHandler(ctx, events.UpdateEducatorCommand{
		ID:         educatorID,
		GivenName:  signals.Educator.GivenName,
		ChosenName: signals.Educator.ChosenName,
		FamilyName: signals.Educator.FamilyName,
		Role:       signals.Educator.Role,
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
	sse.Redirect(fmt.Sprintf("/educators/%s", educatorID))
}

// POST request to /educators/{id}
func (s Server) deleteEducator(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	_, err := events.DeleteEducatorCommandHandler(r.Context(), events.DeleteEducatorCommand{
		EducatorID: educatorID,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}

// HELPER FUNCTIONS
func (s Server) refreshEducatorViewState(ctx context.Context, educatorID string) error {
	educator, err := s.Educators.Get(ctx, educatorID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, educator.ID+".view", educator)
}

func (s Server) refreshEducatorEditState(ctx context.Context, educatorID string) error {
	educator, err := s.Educators.Get(ctx, educatorID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, educator.ID+".edit", educator)
}
