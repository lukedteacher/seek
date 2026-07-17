package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/iep_services/dto"
	"seek/internal/features/iep_services/events"
	"seek/internal/features/iep_services/models"
	"seek/internal/features/iep_services/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) iepServiceRoutes(r chi.Router) {
	r.Get("/iepservices/list", s.getIEPServicesList)
	r.Get("/iepservices/list/stream", s.getIEPServicesListStream)
	r.Get("/iepservices/create", s.getIEPServiceCreate)
	r.Get("/iepservices/create/stream", s.getIEPServiceCreateStream)
	r.Post("/iepservices/create/validate", s.postIEPServiceCreateValidate)
	r.Post("/iepservices/create", s.postIEPServiceCreate)
	r.Get("/iepservices/{id}/view", s.getIEPServiceView)
	r.Get("/iepservices/{id}/view/stream", s.getIEPServiceViewStream)
	r.Get("/iepservices/{id}/edit", s.getIEPServiceEdit)
	r.Get("/iepservices/{id}/edit/stream", s.getIEPServiceEditStream)
	r.Post("/iepservices/{id}/edit", s.postIEPServiceEdit)
	r.Post("/iepservices/{id}/edit/validate", s.postIEPServiceEditValidate)
	r.Delete("/iepservices/{id}/delete", s.deleteIEPService)
}

// GET request to /iepservices
func (s Server) getIEPServicesList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	type Signals struct {
		View int `json:"view"`
	}
	signals := &Signals{}
	datastar.ReadSignals(r, signals)
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("signal read error: ", err.Error())
		return
	}
	iepServices, err := s.IEPServices.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	iepServiceViews := make([]dto.IEPServiceView, len(iepServices))
	for i := range iepServices {
		iepService := dto.NewIEPServiceView(&iepServices[i])
		if err != nil {
			println("error: ", err.Error())
			return
		}
		iepServiceViews[i] = iepService
	}

	_ = pages.List(user, signals.View, iepServiceViews).Render(ctx, w)
}

// GET request to /iepServices/stream
func (s Server) getIEPServicesListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)
	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any iepServices
	sub, err := s.Subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
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
			iepServices, err := s.IEPServices.List(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			iepServiceViews := make([]dto.IEPServiceView, len(iepServices))
			for i := range iepServices {
				iepServiceView := dto.NewIEPServiceView(&iepServices[i])
				if err != nil {
					println("error: ", err.Error())
					return
				}
				iepServiceViews[i] = iepServiceView
			}

			sse.PatchElementTempl(pages.List(user, 0, iepServiceViews))
		}
	}
}

// GET request to /iepServices/create
// TODO figure out if there's a way to have this use the same form as edit?
func (s Server) getIEPServiceCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.NewIEPService()
	view := dto.NewIEPServiceFormView(empty)
	_ = pages.Create(user, view).Render(ctx, w)
}

// GET request to /iepServices/create/stream
func (s Server) getIEPServiceCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		"new",
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
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			println("watcher update")
			if !ok {
				return
			}
			var model models.IEPService
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewIEPServiceFormView(&model)
			sse.PatchElementTempl(pages.Create(user, view))
		}
	}
}

// POST request to /iepServices/create/validate
func (s Server) postIEPServiceCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		View dto.IEPServiceFormView `json:"view"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pcv signal read: ", err.Error())
		return
	}
	model := dto.NewModelFromView(&signals.View.IEPService)
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "new", model); err != nil {
		println("view store error ", err.Error())
	}
}

// POST request to /iepServices/create
func (s Server) postIEPServiceCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		View dto.IEPServiceFormView `json:"view"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pc signal read: ", err.Error())
		return
	}
	model := dto.NewModelFromView(&signals.View.IEPService)
	validation := events.Validate(&model)
	if validation == nil {
		println("some error")
		return
		// TODO actually validate
	}
	createIEPServiceCommand := events.CreateIEPServiceCommand{
		StudentID:   model.StudentID,
		ServiceType: model.ServiceType,
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	_, err := events.CreateIEPServiceCommandHandler(ctx, createIEPServiceCommand, s.EventSaver)
	if err != nil {
		println("ph cpch error: ", err.Error())
		return
	}
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&dto.IEPServiceView{}, sse)
	})
}

// GET request to /iepServices/{id}
func (s Server) getIEPServiceView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	model, err := s.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}

	view := dto.NewIEPServiceView(model)
	if err != nil {
		println("error: ", err.Error())
		return
	}
	_ = pages.View(user, view).Render(ctx, w)
}

// GET request to /iepServices/{id}/stream
func (s Server) getIEPServiceViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
		return
	}
	defer sub.Close()

	if err := s.refreshIEPServiceViewState(ctx, iepServiceID); err != nil {
		println("pvs first refresh: ", err.Error())
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		iepServiceID+".view",
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
			if err := s.refreshIEPServiceViewState(ctx, iepServiceID); err != nil {
				println("pvs second refresh: ", err.Error())
				if err.Error() == "iepService not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			var model models.IEPService
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewIEPServiceView(&model)
			if err != nil {
				println("error: ", err.Error())
			}
			sse.PatchElementTempl(pages.View(user, view))
		}
	}
}

// GET request to /iepServices/{id}/edit
func (s Server) getIEPServiceEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	model, err := s.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewIEPServiceFormView(model)
	_ = pages.Edit(user, view).Render(ctx, w)
}

// GET request to /iepService/{id}/stream
func (s Server) getIEPServiceEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
		return
	}
	defer sub.Close()

	// watches the iepService edit view state kv
	watcher, err := s.ViewStore.Watch(
		ctx,
		iepServiceID+".edit",
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
			if err := s.refreshIEPServiceEditState(ctx, iepServiceID); err != nil {
				println(err.Error())
				if err.Error() == "iepService not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model models.IEPService
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view := dto.NewIEPServiceFormView(&model)
			sse.PatchElementTempl(pages.Edit(user, view))
		}
	}
}

// POST request to /iepServices/{id}/edit/validate
func (s Server) postIEPServiceEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	signals := &struct {
		View dto.IEPServiceFormView `json:"view"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vep signals: ", err.Error())
		return
	}
	signals.View.IEPService.IEPServiceID = iepServiceID
	model := dto.NewModelFromView(&signals.View.IEPService)
	viewstore.PutState(ctx, s.ViewStore, iepServiceID, model)
}

// POST request to /iepServices/{id}/edit
func (s Server) postIEPServiceEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		View dto.IEPServiceFormView `json:"view"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("iep service edit read signals error: ", err.Error())
		return
	}
	iepServiceID := chi.URLParam(r, "id")
	command := events.UpdateIEPServiceCommand{
		IEPServiceID: iepServiceID,
		StudentID:    signals.View.IEPService.StudentID,
		ServiceType:  signals.View.IEPService.ServiceType,
		Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.UpdateIEPServiceCommandHandler(ctx, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(fmt.Errorf("upch error: %w", err))
		return
	}
	if result.Skipped == true {
		println("iep service update skipped")
	}
}

// DELETE request to /iepServices/{id}
func (s Server) deleteIEPService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	_, err := events.DeleteIEPServiceCommandHandler(ctx, events.DeleteIEPServiceCommand{
		IEPServiceID: iepServiceID,
		Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(err.Error())
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/iepservices/list")
}

func (s Server) refreshIEPServiceViewState(ctx context.Context, iepServiceID string) error {
	iepService, err := s.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	if iepService.ArchivedAt != "" {
		println("this iep service was archived")
	}
	return viewstore.PutState(ctx, s.ViewStore, iepService.IEPServiceID+".view", iepService)
}

func (s Server) refreshIEPServiceEditState(ctx context.Context, iepServiceID string) error {
	iepService, err := s.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, iepService.IEPServiceID+".edit", iepService)
}
