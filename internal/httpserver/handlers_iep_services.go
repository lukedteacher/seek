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
	"seek/internal/features/iepservices/dto"
	"seek/internal/features/iepservices/events"
	"seek/internal/features/iepservices/models"
	"seek/internal/features/iepservices/pages"
	sevents "seek/internal/features/students/events"
	studentEvents "seek/internal/features/students/events"
	"seek/internal/ui/core/coreblocks/toasts"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) iepServiceRoutes(r chi.Router) {
	r.Get("/iepservices", getIEPServicesList(s.Logger))
	r.Get("/iepservices/stream", getIEPServicesListStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPServices))
	r.Get("/iepservices/create", getIEPServiceCreate(s.Logger))
	r.Get("/iepservices/create/stream", getIEPServiceCreateStream(s.Logger, s.ViewStore, *s.ReadModels.Students))
	r.Post("/iepservices/create/validate", postIEPServiceCreateValidate(s.Logger, s.ViewStore))
	r.Post("/iepservices/create", postIEPServiceCreate(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/iepservices/{id}", getIEPServiceView(s.Logger))
	r.Get("/iepservices/{id}/stream", getIEPServiceViewStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPServices))
	r.Get("/iepservices/{id}/edit", getIEPServiceEdit(s.Logger))
	r.Get("/iepservices/{id}/edit/stream", getIEPServiceEditStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPServices, *s.ReadModels.Students))
	r.Post("/iepservices/{id}/edit", postIEPServiceEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/iepservices/{id}/edit/validate", postIEPServiceEditValidate(s.Logger, s.ViewStore))
	r.Delete("/iepservices/{id}", deleteIEPService(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/iepservices/csv", getCSV(s.Logger, *s.ReadModels.IEPServices, *s.ReadModels.Students))
	r.Post("/iepservices/csv", postCSV(s.Logger, s.EventSaver, s.EventRetriever, *s.ReadModels.IEPServices, *s.ReadModels.Students))
}

// GET request to /iepservices
func getIEPServicesList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewIEPServiceTableView([]models.IEPService{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /iepservices/stream
func getIEPServicesListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	_ viewstore.Store,
	serviceReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any iepServices
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep services list stream subscribe", "err", err)
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
				services, err := serviceReadModel.List(ctx)
				if err != nil {
					l.ErrorContext(ctx, "iep services list stream db list", "err", err)
					return
				}
				view := dto.NewIEPServiceTableView(services)
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /iepservices/create
func getIEPServiceCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Create(dto.IEPServiceFormView{FormType: "create"}).Render(ctx, w)
	}
}

// GET request to /iepservices/create/stream
func getIEPServiceCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	studentReadModel studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			"new",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep create stream watcher init", "err", err)
			return
		}
		defer watcher.Stop()

		students, err := studentReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "service create stream list students", "err", err)
		}
		view := dto.NewIEPServiceFormView(
			"create",
			&models.IEPService{},
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
				model := &models.IEPService{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep create stream watcher update", "err", err)
					return
				}
				students, _ := studentReadModel.List(ctx)
				view := dto.NewIEPServiceFormView(
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

// POST request to /iepservices/create/validate
func postIEPServiceCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			View dto.IEPServiceView `json:"iepservice"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "iep create validate signals read", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		// saves the state to a view store so that the SSE can update
		// TODO look into a better name for the channel
		if err := viewstore.PutState(ctx, vs, "new", model); err != nil {
			l.ErrorContext(ctx, "post iep services create validate viewstore", "err", err)
		}
	}
}

// POST request to /iepservices/create
func postIEPServiceCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			View dto.IEPServiceView `json:"iepservice"`
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
			StudentID:       model.StudentID,
			ServiceType:     signals.View.ServiceType.ShortString(),
			IndirectMinutes: model.IndirectMinutes,
			DirectMinutes:   model.DirectMinutes,
			FrequencyCount:  model.FrequencyCount,
			FrequencyType:   model.FrequencyType,
			LocationID:      model.LocationID,
			Provider:        model.Provider,
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
		sse.Redirect(fmt.Sprintf("/iepservices/%s", result.EventID))
	}
}

// GET request to /iepservices/{id}
func getIEPServiceView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewIEPServiceView(&models.IEPService{})
		_ = pages.View(view).Render(ctx, w)
	}
}

// GET request to /iepservices/{id}/stream
func getIEPServiceViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepServiceID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
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
			iepServiceID+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshIEPServiceViewState(ctx, l, vs, iepServiceID, serviceReadModel); err != nil {
			l.ErrorContext(ctx, "iep service view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshIEPServiceViewState(ctx, l, vs, iepServiceID, serviceReadModel); err != nil {
					if err.Error() == "iepService not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				model := &models.IEPService{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service view stream json", "err", err)
					return
				}
				view := dto.NewIEPServiceView(model)
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

// GET request to /iepservices/{id}/edit
func getIEPServiceEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Edit(dto.IEPServiceFormView{FormType: "edit"}).Render(ctx, w)
	}
}

// GET request to /iepService/{id}/stream
func getIEPServiceEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel events.ReadModel,
	studentReadModel studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepServiceID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(iepServiceID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the iepService edit view state kv
		watcher, err := vs.Watch(
			ctx,
			iepServiceID+".edit",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshIEPServiceEditState(ctx, l, vs, iepServiceID, serviceReadModel); err != nil {
			if err.Error() == "iepService not found" {
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
				if err := refreshIEPServiceEditState(ctx, l, vs, iepServiceID, serviceReadModel); err != nil {
					if err.Error() == "iepService not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service edit stream refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.IEPService{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service edit stream json", "err", err)
					return
				}
				students, _ := studentReadModel.List(ctx)
				view := dto.NewIEPServiceFormView(
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

// POST request to /iepservices/{id}/edit/validate
func postIEPServiceEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepServiceID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.IEPServiceView `json:"iepservice"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit validate signals", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		viewstore.PutState(ctx, vs, iepServiceID, model)
	}
}

// POST request to /iepservices/{id}/edit
func postIEPServiceEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		iepServiceID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.IEPServiceView `json:"iepservice"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit signals", "err", err)
			return
		}
		command := events.UpdateIEPServiceCommand{
			IEPServiceID:    iepServiceID,
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
		result, err := events.UpdateIEPServiceCommandHandler(ctx, command, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post iep service edit command handler", "err", err)
			return
		}
		if result.Skipped == true {
			l.Info("post iep service edit command handler", "skipped", result.Skipped)
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/iepservices/%s", iepServiceID))
	}
}

// DELETE request to /iepservices/{id}
func deleteIEPService(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		iepServiceID := chi.URLParam(r, "id")
		_, err := events.DeleteIEPServiceCommandHandler(ctx, events.DeleteIEPServiceCommand{
			IEPServiceID: iepServiceID,
			Metadata:     eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete iep service command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/iepservices")
	}
}

// GET request to /iepservices/csv
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

		// get student list and make map for marss check
		students, _ := studentReadModel.List(ctx)
		marssMap := make(map[string]string)
		for _, student := range students {
			marssMap[student.MARSSID] = student.ID
		}

		// convert CSV rows with valid MARSS ID to domain models
		converted := make([]*models.IEPService, 0)
		for _, csvSvc := range filtered {
			key := strings.TrimSpace(csvSvc.StudentMARSSID)
			if student_id, ok := marssMap[key]; ok {
				csvSvc.StudentID = student_id
				converted = append(converted, csvSvc.ToIEPService())
			}
		}

		// fetch existing DB services
		dbServices, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "read csv list db services", "err", err)
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
		pages.ReadCSV(view).Render(ctx, w)
	}
}

// POST request to /iepservices/csv
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

		// get student list and make map for marss check
		students, _ := studentReadModel.List(ctx)
		marssMap := make(map[string]string)
		for _, student := range students {
			marssMap[student.MARSSID] = student.ID
		}

		// convert CSV rows with valid MARSS ID to domain models
		converted := make([]*models.IEPService, 0)
		for _, csvSvc := range filtered {
			key := strings.TrimSpace(csvSvc.StudentMARSSID)
			if student_id, ok := marssMap[key]; ok {
				csvSvc.StudentID = student_id
				converted = append(converted, csvSvc.ToIEPService())
			}
		}

		// fetch existing DB services
		dbServices, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "read csv list db services", "err", err)
			return
		}
		dbPtrs := make([]*models.IEPService, len(dbServices))
		for i := range dbServices {
			dbPtrs[i] = &dbServices[i]
		}

		// compute diff
		diffs := models.CompareIEPServices(dbPtrs, converted)

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
						Provider:        diff.New.Provider,
					},
					saver,
					retriever,
				)
			}
			if diff.Status == sharedmodels.DiffUpdated {
				events.UpdateIEPServiceCommandHandler(
					ctx,
					events.UpdateIEPServiceCommand{
						IEPServiceID:    diff.New.ID,
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
				events.DeleteIEPServiceCommandHandler(
					ctx,
					events.DeleteIEPServiceCommand{
						IEPServiceID: diff.Old.ID,
						StudentID:    diff.Old.StudentID,
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
		sse.Redirect("/iepservices")
	}
}

func refreshIEPServiceViewState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	iepServiceID string,
	serviceReadModel events.ReadModel,
) error {
	iepService, err := serviceReadModel.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, iepService.ID+".view", iepService)
}

func refreshIEPServiceEditState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	iepServiceID string,
	serviceReadModel events.ReadModel,
) error {
	iepService, err := serviceReadModel.Get(ctx, iepServiceID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, iepService.ID+".edit", iepService)
}
