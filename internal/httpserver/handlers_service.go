package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/services/dto"
	"seek/internal/features/services/events"
	"seek/internal/features/services/models"
	"seek/internal/features/services/pages"
	sevents "seek/internal/features/students/events"
	studentEvents "seek/internal/features/students/events"
	"seek/internal/ui/core/coreblocks/toasts"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) serviceRoutes(r chi.Router) {
	r.Get("/services", getServicesList(s.Logger))
	r.Get("/services/stream", getServicesListStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Services))
	r.Get("/services/create", getServiceCreate(s.Logger))
	r.Get("/services/create/stream", getServiceCreateStream(s.Logger, s.ViewStore, *s.ReadModels.Students))
	r.Post("/services/create/validate", postServiceCreateValidate(s.Logger, s.ViewStore))
	r.Post("/services/create", postServiceCreate(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/services/{id}", getServiceView(s.Logger))
	r.Get("/services/{id}/stream", getServiceViewStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Services))
	r.Get("/services/{id}/edit", getServiceEdit(s.Logger))
	r.Get("/services/{id}/edit/stream", getServiceEditStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.Services, *s.ReadModels.Students))
	r.Post("/services/{id}/edit", postServiceEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/services/{id}/edit/validate", postServiceEditValidate(s.Logger, s.ViewStore))
	r.Delete("/services/{id}", deleteService(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/services/csv", getCSV(s.Logger, *s.ReadModels.Services, *s.ReadModels.Students))
	r.Post("/services/csv", postCSV(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.Services, *s.ReadModels.Students))
}

// GET request to /services
func getServicesList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewServiceTableView([]models.Service{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /services/stream
func getServicesListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	_ viewstore.Store,
	serviceReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any services
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep services list stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		services, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "iep services list stream db list", "err", err)
			return
		}
		view := dto.NewServiceTableView(services)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				// for now just refreshes the page
				// consider adding a view store for the list
				services, err := serviceReadModel.List(ctx)
				if err != nil {
					l.ErrorContext(ctx, "iep services list stream db list", "err", err)
					return
				}
				view := dto.NewServiceTableView(services)
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /services/create
func getServiceCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Create(dto.ServiceFormView{FormType: "create"}).Render(ctx, w)
	}
}

// GET request to /services/create/stream
func getServiceCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	studentReadModel studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			user.Username+".services.create",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep create stream watcher init", "err", err)
			return
		}
		defer watcher.Stop()

		students, err := studentReadModel.ListWithIEPs(ctx)
		if err != nil {
			l.ErrorContext(ctx, "service create stream list students", "err", err)
		}
		view := dto.NewServiceFormView(
			"create",
			&models.Service{},
			students,
			[]educatorModels.Educator{},
		)
		sse.PatchElementTempl(pages.Create(view))

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				model := &models.Service{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep create stream watcher update", "err", err)
					return
				}
				students, _ := studentReadModel.ListWithIEPs(ctx)
				view := dto.NewServiceFormView(
					"create",
					model,
					students,
					[]educatorModels.Educator{},
				)
				sse.PatchElementTempl(pages.Create(view))
			}
		}
	}
}

// POST request to /services/create/validate
func postServiceCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			View dto.ServiceView `json:"service"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "iep create validate signals read", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		// saves the state to a view store so that the SSE can update
		// TODO look into a better name for the channel
		if err := viewstore.PutState(ctx, vs, user.Username+".services.create", model); err != nil {
			l.ErrorContext(ctx, "post iep services create validate viewstore", "err", err)
		}
	}
}

// POST request to /services/create
func postServiceCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			View dto.ServiceView `json:"service"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep services create signals", "err", err)
			return
		}
		if signals.View.StudentID == "" {
			sse := newSSE(w, r)
			sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantError, "no student selected"))
			return
		}
		model := dto.NewModelFromView(&signals.View)
		cmd := events.AddServiceToIEPCommand{
			IEPID:           model.IEPID,
			StudentID:       model.StudentID,
			ServiceType:     signals.View.ServiceType.ShortString(),
			IndirectMinutes: model.IndirectMinutes,
			DirectMinutes:   model.DirectMinutes,
			FrequencyCount:  model.FrequencyCount,
			FrequencyType:   model.FrequencyType,
			LocationID:      model.LocationID,
			ProviderID:      model.ProviderID,
			StartDate:       model.StartDate.String(),
			EndDate:         model.EndDate.String(),
			Metadata:        eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.AddServiceToIEPCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post iep services create command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/services/%s", result.EventID))
	}
}

// GET request to /services/{id}
func getServiceView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewServiceView(&models.Service{})
		_ = pages.View(view).Render(ctx, w)
	}
}

// GET request to /services/{id}/stream
func getServiceViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		serviceID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(serviceID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep service view stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			serviceID+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshServiceViewState(ctx, l, vs, serviceID, serviceReadModel); err != nil {
			l.ErrorContext(ctx, "iep service view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshServiceViewState(ctx, l, vs, serviceID, serviceReadModel); err != nil {
					if err.Error() == "service not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				model := &models.Service{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service view stream json", "err", err)
					return
				}
				view := dto.NewServiceView(model)
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

// GET request to /services/{id}/edit
func getServiceEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Edit(dto.ServiceFormView{FormType: "edit"}).Render(ctx, w)
	}
}

// GET request to /service/{id}/stream
func getServiceEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel events.ReadModel,
	studentReadModel studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		serviceID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(serviceID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the service edit view state kv
		watcher, err := vs.Watch(
			ctx,
			serviceID+".edit",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshServiceEditState(ctx, l, vs, serviceID, serviceReadModel); err != nil {
			if err.Error() == "service not found" {
				sse.PatchElementTempl(pages.NotFound())
			}
			l.ErrorContext(ctx, "service edit stream refresh state", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				if err := refreshServiceEditState(ctx, l, vs, serviceID, serviceReadModel); err != nil {
					if err.Error() == "service not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service edit stream refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.Service{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service edit stream json", "err", err)
					return
				}
				students, _ := studentReadModel.ListWithIEPs(ctx)
				view := dto.NewServiceFormView(
					"edit",
					model,
					students,
					[]educatorModels.Educator{},
				)
				sse.PatchElementTempl(pages.Edit(view))
			}
		}
	}
}

// POST request to /services/{id}/edit/validate
func postServiceEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		serviceID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.ServiceView `json:"service"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit validate signals", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		viewstore.PutState(ctx, vs, serviceID, model)
	}
}

// POST request to /services/{id}/edit
func postServiceEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		serviceID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.ServiceView `json:"service"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit signals", "err", err)
			return
		}
		cmd := events.UpdateServiceCommand{
			ServiceID:       serviceID,
			IEPID:           signals.View.IEPID,
			StudentID:       signals.View.StudentID,
			ServiceName:     signals.View.ServiceName,
			ServiceType:     signals.View.ServiceType.ShortString(),
			IndirectMinutes: signals.View.IndirectMinutes,
			DirectMinutes:   signals.View.DirectMinutes,
			FrequencyCount:  signals.View.FrequencyCount,
			FrequencyType:   signals.View.FrequencyType,
			LocationID:      signals.View.LocationID,
			StartDate:       signals.View.StartDate.String(),
			EndDate:         signals.View.EndDate.String(),
			ProviderID:      signals.View.ProviderID,
			Metadata:        eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateServiceCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post iep service edit command handler", "err", err)
			return
		}
		if result.Skipped == true {
			l.Info("post iep service edit command handler", "skipped", result.Skipped)
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/services/%s", serviceID))
	}
}

// DELETE request to /services/{id}
func deleteService(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		serviceID := chi.URLParam(r, "id")
		_, err := events.DeleteServiceCommandHandler(ctx, events.DeleteServiceCommand{
			ServiceID: serviceID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete iep service command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/services")
	}
}

// GET request to /services/csv
func getCSV(
	l *slog.Logger,
	serviceReadModel events.ReadModel,
	studentReadModel sevents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, err := os.OpenFile("iep_services.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			http.Error(w, "failed to open csv file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		csvServices := []models.CSVService{}
		if err := gocsv.UnmarshalFile(file, &csvServices); err != nil {
			http.Error(w, "failed to parse csv: "+err.Error(), http.StatusBadRequest)
			return
		}

		// filter out unwanted services
		filtered := make([]models.CSVService, 0, len(csvServices))
		for _, svc := range csvServices {
			if svc.ServiceName != "Shared paraprofessional" {
				filtered = append(filtered, svc)
			}
		}

		// get student list and make map for marss check
		students, _ := studentReadModel.List(ctx)
		marssMap := make(map[string]string)
		for _, student := range students {
			marssMap[student.MARSSID] = student.ID
		}

		// convert CSV rows with valid MARSS ID to domain models
		converted := make([]models.Service, 0)
		for _, csvSvc := range filtered {
			key := strings.TrimSpace(csvSvc.StudentMARSSID)
			if student_id, ok := marssMap[key]; ok {
				csvSvc.StudentID = student_id
				converted = append(converted, csvSvc.ToService())
			}
		}

		// fetch existing DB services
		dbServices, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "read csv list db services", "err", err)
			return
		}

		// compute diff
		diffs := models.CompareServices(dbServices, converted)

		// render view
		view := dto.NewServiceDiffTableView(diffs)
		pages.ReadCSV(view).Render(ctx, w)
	}
}

// POST request to /services/csv
func postCSV(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	serviceReadModel events.ReadModel,
	studentReadModel sevents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, err := os.OpenFile("iep_services.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			http.Error(w, "failed to open csv file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		csvServices := []*models.CSVService{}
		if err := gocsv.UnmarshalFile(file, &csvServices); err != nil {
			http.Error(w, "failed to parse csv: "+err.Error(), http.StatusBadRequest)
			return
		}

		// filter out unwanted services
		filtered := make([]*models.CSVService, 0, len(csvServices))
		for _, svc := range csvServices {
			if svc.ServiceName != "Shared paraprofessional" {
				filtered = append(filtered, svc)
			}
		}

		// get student list and make map for marss check
		students, _ := studentReadModel.List(ctx)
		marssMap := make(map[string]string)
		for _, student := range students {
			marssMap[student.MARSSID] = student.ID
		}

		// convert CSV rows with valid MARSS ID to domain models
		converted := make([]models.Service, 0)
		for _, csvSvc := range filtered {
			key := strings.TrimSpace(csvSvc.StudentMARSSID)
			if student_id, ok := marssMap[key]; ok {
				csvSvc.StudentID = student_id
				converted = append(converted, csvSvc.ToService())
			}
		}

		// fetch existing DB services
		dbServices, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "read csv list db services", "err", err)
			return
		}

		// compute diff
		diffs := models.CompareServices(dbServices, converted)

		for _, diff := range diffs {
			if diff.Status == sharedmodels.DiffNew {
				events.AddServiceToIEPCommandHandler(
					ctx,
					events.AddServiceToIEPCommand{
						StudentID:       diff.New.StudentID,
						ServiceName:     diff.New.ServiceName,
						ServiceType:     string(diff.New.ServiceType),
						IndirectMinutes: diff.New.IndirectMinutes,
						DirectMinutes:   diff.New.DirectMinutes,
						FrequencyCount:  diff.New.FrequencyCount,
						FrequencyType:   diff.New.FrequencyType,
						StartDate:       diff.New.StartDate.String(),
						EndDate:         diff.New.EndDate.String(),
						LocationID:      diff.New.LocationID,
						ProviderID:      diff.New.ProviderID,
					},
					saver,
					retriever,
				)
			}
			if diff.Status == sharedmodels.DiffUpdated {
				events.UpdateServiceCommandHandler(
					ctx,
					events.UpdateServiceCommand{
						ServiceID:       diff.New.ID,
						StudentID:       diff.New.StudentID,
						ServiceName:     diff.New.ServiceName,
						ServiceType:     string(diff.New.ServiceType),
						IndirectMinutes: diff.New.IndirectMinutes,
						DirectMinutes:   diff.New.DirectMinutes,
						FrequencyCount:  diff.New.FrequencyCount,
						FrequencyType:   diff.New.FrequencyType,
						StartDate:       diff.New.StartDate.String(),
						EndDate:         diff.New.EndDate.String(),
						LocationID:      diff.New.LocationID,
						ProviderID:      diff.New.ProviderID,
					},
					saver,
					retriever,
				)
			}
			if diff.Status == sharedmodels.DiffAbsent {
				events.DeleteServiceCommandHandler(
					ctx,
					events.DeleteServiceCommand{
						ServiceID: diff.Old.ID,
						StudentID: diff.Old.StudentID,
					},
					saver,
					retriever,
				)
			}
			if diff.Status == sharedmodels.DiffSame {
				continue
			}
		}

		sse := newSSE(w, r)
		sse.Redirect("/services")
	}
}

func refreshServiceViewState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	serviceID string,
	serviceReadModel events.ReadModel,
) error {
	service, err := serviceReadModel.Get(ctx, serviceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, service.ID+".view", service)
}

func refreshServiceEditState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	serviceID string,
	serviceReadModel events.ReadModel,
) error {
	service, err := serviceReadModel.Get(ctx, serviceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, service.ID+".edit", service)
}
