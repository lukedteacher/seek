package httpserver

import (
	"context"
	"log"
	"net/http"

	"seek/internal/eventstore"
	sdto "seek/internal/features/_shared/dto"
	sm "seek/internal/features/_shared/models"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/events"
	"seek/internal/features/educators/models"
	"seek/internal/features/educators/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) educatorRoutes(r chi.Router) {
	r.Get("/educators/list", s.getEducatorsList)
	r.Get("/educators/list/stream", s.getEducatorsListStream)
	r.Get("/educators/create", s.getEducatorCreate)
	r.Get("/educators/create/stream", s.getEducatorCreateStream)
	r.Post("/educators/create/validate", s.postEducatorCreateValidate)
	r.Post("/educators/create", s.postEducatorCreate)
	r.Get("/educators/{id}/view", s.getEducatorView)
	r.Get("/educators/{id}/view/stream", s.getEducatorViewStream)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := sdto.BuildTableView(educators, nil, []string{
		"GivenName",
		"ChosenName",
		"FamilyName",
		"Role",
		"Email",
	})
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
		println("educators list stream error: ", err.Error())
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
			view := sdto.BuildTableView(educators, nil, []string{
				"GivenName",
				"ChosenName",
				"FamilyName",
				"Role",
				"Email",
			})

			sse.PatchElementTempl(pages.List(user, view))
		}
	}
}

// GET request to /educators/create
func (s Server) getEducatorCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.NewEducator()
	view := dto.NewEducatorView(empty)
	_ = pages.Create(user, *view).Render(ctx, w)
}

// GET request to /educators/create/stream
func (s Server) getEducatorCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		"neweducator",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		println("watcher error in educator create stream: ", err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			println("educator watcher update")
			if !ok {
				return
			}
			var model models.Educator
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewEducatorView(&model)
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
		println("pcv signal read: ", err.Error())
		return
	}
	model := models.Educator{
		ID: signals.Educator.ID,
		Person: sm.Person{
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
		},
		Role:  signals.Educator.Role,
		Email: signals.Educator.Email,
	}
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "neweducator", model); err != nil {
		println("view store error ", err.Error())
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
		println(err.Error())
		return
	}
	println("seg:", signals.Educator.GivenName)
	_, err := events.CreateEducatorCommandHandler(ctx, events.CreateEducatorCommand{
		GivenName:  signals.Educator.GivenName,
		ChosenName: signals.Educator.ChosenName,
		FamilyName: signals.Educator.FamilyName,
		Role:       signals.Educator.Role,
		Email:      signals.Educator.Email,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

// GET request to /educators/{id}
func (s Server) getEducatorView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	educatorID := chi.URLParam(r, "id")
	educator, err := s.Educators.Get(ctx, educatorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewEducatorView(educator)
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
		println(err.Error())
		return
	}
	defer sub.Close()

	if err := s.refreshEducatorViewState(ctx, educatorID); err != nil {
		println("svs first refresh: ", err.Error())
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
		println(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshEducatorViewState(ctx, educatorID); err != nil {
				println("svs second refresh: ", err.Error())
				if err.Error() == "educator not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			var model models.Educator
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewEducatorView(&model)
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
		println(err.Error())
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
		println(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshEducatorEditState(ctx, educatorID); err != nil {
				println(err.Error())
				if err.Error() == "educator not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model models.Educator
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewEducatorView(&model)
			sse.PatchElementTempl(pages.Edit(user, *view))
		}
	}
}

// POST request to /educators/{id}/edit/validate
func (s Server) postEducatorEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.Educator{
		ID:     signals.Educator.ID,
		Person: signals.Educator.Person,
		Role:   signals.Educator.Role,
		Email:  signals.Educator.Email,
	}
	model.ID = chi.URLParam(r, "id")
	view := dto.NewEducatorView(&model)
	_ = pages.Edit(user, *view).Render(ctx, w)
}

// POST request to /educators/{id}/edit
func (s Server) postEducatorEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Educator dto.EducatorView `json:"educator"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("error reading signals: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		println("command error: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Skipped == true {
		println("update skipped")
		return
	}
}

// POST request to /educators/{id}/delete
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
