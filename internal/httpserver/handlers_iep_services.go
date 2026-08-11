package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"seek/internal/eventstore"
	"seek/internal/features/iepservices/dto"
	"seek/internal/features/iepservices/events"
	"seek/internal/features/iepservices/models"
	"seek/internal/features/iepservices/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) iepServiceRoutes(r chi.Router) {
	r.Get("/iepservices", s.getIEPServicesList)
	r.Get("/iepservices/stream", s.getIEPServicesListStream)
	r.Get("/iepservices/create", s.getIEPServiceCreate)
	r.Get("/iepservices/create/stream", s.getIEPServiceCreateStream)
	r.Post("/iepservices/create/validate", s.postIEPServiceCreateValidate)
	r.Post("/iepservices/create", s.postIEPServiceCreate)
	r.Get("/iepservices/{id}", s.getIEPServiceView)
	r.Get("/iepservices/{id}/stream", s.getIEPServiceViewStream)
	r.Get("/iepservices/{id}/edit", s.getIEPServiceEdit)
	r.Get("/iepservices/{id}/edit/stream", s.getIEPServiceEditStream)
	r.Post("/iepservices/{id}/edit", s.postIEPServiceEdit)
	r.Post("/iepservices/{id}/edit/validate", s.postIEPServiceEditValidate)
	r.Delete("/iepservices/{id}", s.deleteIEPService)
	r.Get("/iepservices/csv", s.ReadCSV)
}

// GET request to /iepservices
func (s Server) getIEPServicesList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services, err := s.ReadModels.IEPServices.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep services list db list", "err", err)
		return
	}
	view := dto.NewIEPServiceTableView(services)
	_ = pages.List(view).Render(ctx, w)
}

// GET request to /iepservices/stream
func (s Server) getIEPServicesListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sse := newSSE(w, r)
	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any iepServices
	sub, err := s.Subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep services list stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			// for now just refreshes the page
			// consider adding a view store for the list
			services, err := s.ReadModels.IEPServices.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "iep services list stream db list", "err", err)
				return
			}
			view := dto.NewIEPServiceTableView(services)
			sse.PatchElementTempl(pages.List(view))
		}
	}
}

// GET request to /iepservices/create
// TODO figure out if there's a way to have this use the same form as edit?
func (s Server) getIEPServiceCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	empty := models.NewIEPService()
	students, _ := s.ReadModels.Students.List(ctx)
	view := dto.NewIEPServiceFormView(empty, students)
	_ = pages.Create(view).Render(ctx, w)
}

// GET request to /iepservices/create/stream
func (s Server) getIEPServiceCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
		s.Logger.ErrorContext(ctx, "iep create stream watcher init", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.IEPService{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "iep create stream watcher update", "err", err)
				return
			}
			students, _ := s.ReadModels.Students.List(ctx)
			view := dto.NewIEPServiceFormView(model, students)
			sse.PatchElementTempl(pages.Create(view))
		}
	}
}

// POST request to /iepservices/create/validate
func (s Server) postIEPServiceCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		View dto.IEPServiceView `json:"iepservice"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "iep create validate signals read", "err", err)
		return
	}
	model := dto.NewModelFromView(&signals.View)
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "new", model); err != nil {
		s.Logger.ErrorContext(ctx, "post iep services create validate viewstore", "err", err)
	}
}

// POST request to /iepservices/create
func (s Server) postIEPServiceCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		View dto.IEPServiceView `json:"iepservice"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post iep services create signals", "err", err)
		return
	}
	if signals.View.StudentID == "" {
		s.Logger.InfoContext(ctx, "no student ID!")
		return
	}
	model := dto.NewModelFromView(&signals.View)
	command := events.AddIEPServiceToStudentCommand{
		StudentID:       model.StudentID,
		ServiceType:     signals.View.ServiceType.ShortString(),
		IndirectMinutes: model.IndirectMinutes,
		DirectMinutes:   model.DirectMinutes,
		FrequencyCount:  model.FrequencyCount,
		FrequencyType:   model.FrequencyType,
		Location:        model.Location,
		Provider:        model.Provider,
		StartDate:       model.StartDate.String(),
		EndDate:         model.EndDate.String(),
		Metadata:        eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.AddIEPServiceToStudentCommandHandler(ctx, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post iep services create command handler", "err", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/iepservices/%s", result.EventID))
}

// GET request to /iepservices/{id}
func (s Server) getIEPServiceView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	model, err := s.ReadModels.IEPServices.Get(ctx, iepServiceID)
	if model == nil {
		_ = pages.NotFound().Render(ctx, w)
		return
	}
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep service view db get", "err", err)
		return
	}

	view := dto.NewIEPServiceView(model)
	_ = pages.View(view).Render(ctx, w)
}

// GET request to /iepservices/{id}/stream
func (s Server) getIEPServiceViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep service view stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshIEPServiceViewState(ctx, iepServiceID); err != nil {
		s.Logger.ErrorContext(ctx, "iep service view stream refresh", "err", err)
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
		s.Logger.ErrorContext(ctx, "iep service view stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshIEPServiceViewState(ctx, iepServiceID); err != nil {
				if err.Error() == "iepService not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				s.Logger.ErrorContext(ctx, "iep service view stream refresh in select", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.IEPService{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "iep service view stream json", "err", err)
				return
			}
			view := dto.NewIEPServiceView(model)
			sse.PatchElementTempl(pages.View(view))
		}
	}
}

// GET request to /iepservices/{id}/edit
func (s Server) getIEPServiceEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	model, err := s.ReadModels.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep service edit db get", "err", err)
		return
	}
	students, _ := s.ReadModels.Students.List(ctx)
	view := dto.NewIEPServiceFormView(model, students)
	_ = pages.Edit(view).Render(ctx, w)
}

// GET request to /iepService/{id}/stream
func (s Server) getIEPServiceEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "iep service edit stream subscribe", "err", err)
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
		s.Logger.ErrorContext(ctx, "iep service edit stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshIEPServiceEditState(ctx, iepServiceID); err != nil {
				if err.Error() == "iepService not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				s.Logger.ErrorContext(ctx, "iep service edit stream refresh", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			model := &models.IEPService{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "iep service edit stream json", "err", err)
				return
			}
			students, _ := s.ReadModels.Students.List(ctx)
			view := dto.NewIEPServiceFormView(model, students)
			sse.PatchElementTempl(pages.Edit(view))
		}
	}
}

// POST request to /iepservices/{id}/edit/validate
func (s Server) postIEPServiceEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	iepServiceID := chi.URLParam(r, "id")
	signals := &struct {
		View dto.IEPServiceView `json:"iepservice"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post iep service edit validate signals", "err", err)
		return
	}
	signals.View.ID = iepServiceID
	model := dto.NewModelFromView(&signals.View)
	viewstore.PutState(ctx, s.ViewStore, iepServiceID, model)
}

// POST request to /iepservices/{id}/edit
func (s Server) postIEPServiceEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	signals := &struct {
		View dto.IEPServiceView `json:"iepservice"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post iep service edit signals", "err", err)
		return
	}
	command := events.UpdateIEPServiceCommand{
		IEPServiceID:    iepServiceID,
		StudentID:       signals.View.StudentID,
		ServiceType:     signals.View.ServiceType.ShortString(),
		IndirectMinutes: signals.View.IndirectMinutes,
		DirectMinutes:   signals.View.DirectMinutes,
		FrequencyCount:  signals.View.FrequencyCount,
		FrequencyType:   signals.View.FrequencyType,
		Location:        signals.View.Location,
		StartDate:       signals.View.StartDate.String(),
		EndDate:         signals.View.EndDate.String(),
		Provider:        signals.View.Provider,
		Metadata:        eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.UpdateIEPServiceCommandHandler(ctx, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post iep service edit command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.Info("post iep service edit command handler", "skipped", result.Skipped)
	}
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/iepservices/%s", iepServiceID))
}

// DELETE request to /iepservices/{id}
func (s Server) deleteIEPService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	iepServiceID := chi.URLParam(r, "id")
	_, err := events.DeleteIEPServiceCommandHandler(ctx, events.DeleteIEPServiceCommand{
		IEPServiceID: iepServiceID,
		Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete iep service command handler", "err", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/iepservices")
}

// GET request to /iepservices/csv
func (s Server) ReadCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, err := os.OpenFile("iep_services.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		http.Error(w, "failed to open csv file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	csvServices := []*models.CSVIEPService{}
	if err := gocsv.UnmarshalFile(file, &csvServices); err != nil {
		http.Error(w, "failed to parse csv: "+err.Error(), http.StatusBadRequest)
		return
	}

	// filter out unwanted services
	filtered := make([]*models.CSVIEPService, 0, len(csvServices))
	for _, svc := range csvServices {
		if svc.ServiceName != "Shared paraprofessional" {
			filtered = append(filtered, svc)
		}
	}

	// convert CSV models to domain models
	converted := make([]*models.IEPService, len(filtered))
	for i, csvSvc := range filtered {
		domain := csvSvc.ToIEPService()
		converted[i] = &domain
	}

	// fetch existing DB services
	dbServices, err := s.ReadModels.IEPServices.List(ctx)
	if err != nil {
		http.Error(w, "failed to list existing services: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dbPtrs := make([]*models.IEPService, len(dbServices))
	for i := range dbServices {
		dbPtrs[i] = &dbServices[i]
	}

	// compute diff
	diffs := models.CompareIEPServices(dbPtrs, converted)

	// render view
	view := dto.NewIEPServiceDiffTableView(diffs)
	pages.CSV(view).Render(ctx, w)
}

func (s Server) refreshIEPServiceViewState(ctx context.Context, iepServiceID string) error {
	iepService, err := s.ReadModels.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, iepService.ID+".view", iepService)
}

func (s Server) refreshIEPServiceEditState(ctx context.Context, iepServiceID string) error {
	iepService, err := s.ReadModels.IEPServices.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, iepService.ID+".edit", iepService)
}
