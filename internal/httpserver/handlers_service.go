package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	educatorEvents "seek/internal/features/educators/events"
	iepEvents "seek/internal/features/ieps/events"
	"seek/internal/features/services/dto"
	"seek/internal/features/services/events"
	"seek/internal/features/services/models"
	"seek/internal/features/services/pages"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	studentModels "seek/internal/features/students/models"
	"seek/internal/ui/core/coreblocks/toasts"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) serviceRoutes(r chi.Router) {
	// service list
	r.Get("/services", getServicesList(s.Logger))
	r.Get("/services/stream", getServicesListStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Services))
	r.Get("/servicegrid", getServiceGrid(s.Logger))
	r.Get("/servicegrid/stream", getServiceGridStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Services, s.ReadModels.IEPs, s.ReadModels.Students))
	r.Query("/servicegrid", queryServiceGrid(s.Logger, s.ViewStore))
	// service creation
	r.Get("/services/create", getServiceCreate(s.Logger))
	r.Get("/services/create/stream", getServiceCreateStream(s.Logger, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Students))
	r.Query("/services/create/validate", queryServiceFormValidate(s.Logger, s.ViewStore))
	r.Query("/services/create/{field}/{value}", queryServiceFormField(s.Logger, s.ViewStore))
	r.Post("/services/create", postServiceCreate(s.Logger, s.EventSaver, s.EventRetriever))
	// service viewing
	r.Get("/services/{id}", getServiceView(s.Logger))
	r.Get("/services/{id}/stream", getServiceViewStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Services))
	// service editing
	r.Get("/services/{id}/edit", getServiceEdit(s.Logger))
	r.Get("/services/{id}/edit/stream", getServiceEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Services, s.ReadModels.IEPs, s.ReadModels.Students))
	r.Query("/services/{id}/edit/validate", queryServiceFormValidate(s.Logger, s.ViewStore))
	r.Query("/services/{id}/edit/{field}/{value}", queryServiceFormField(s.Logger, s.ViewStore))
	r.Post("/services/{id}/edit", postServiceEdit(s.Logger, s.EventSaver, s.EventRetriever))
	// other service stuff
	r.Delete("/services/{id}", deleteService(s.Logger, s.EventSaver, s.EventRetriever))
	// service CSV processing
	r.Get("/services/csv", getServicesCSV(s.Logger, s.ReadModels.Services, s.ReadModels.IEPs, s.ReadModels.Students))
	r.Post("/services/csv", postServicesCSV(s.Logger, s.EventSaver, s.EventRetriever, s.ReadModels.Services, s.ReadModels.IEPs, s.ReadModels.Students))
}

// GET request to /services
// renders an empty table template
// SSE will populate data
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
// populates data and keeps it updated with changes pushed from server
func getServicesListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	_ viewstore.Store,
	serviceReadModel *events.ReadModel,
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
			l.ErrorContext(ctx, "gsls sub", "err", err)
			return
		}
		defer sub.Close()

		// initialize data
		services, err := serviceReadModel.List(ctx)
		if err != nil {
			l.ErrorContext(ctx, "iep services list stream db list", "err", err)
			return
		}

		// make the view and push it via SSE
		view := dto.NewServiceTableView(services)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
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

// GET request to /services/grid
// renders an empty grid template
// SSE will populate data
func getServiceGrid(
	l *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Grid(dto.ServiceGrid{}).Render(ctx, w)
	}
}

// GET request to /services/grid/stream
// populates the grid and keeps it updated with read model changes
// and per-user column visibility changes
func getServiceGridStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel *events.ReadModel,
	iepReadModel *iepEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// subscribe to read model changes
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "gsgs sub", "err", err)
			return
		}
		defer sub.Close()

		// watch per-user column visibility
		key := user.Username + ".servicegrid"
		watcher, err := vs.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			l.ErrorContext(ctx, "gsgs watch", "err", err)
			return
		}
		defer watcher.Stop()

		// initial columns: user's saved list, or defaults
		columns := sharedmodels.ServiceTypeList
		if saved, ok, err := viewstore.GetState[dto.ServiceGrid](ctx, vs, key); err != nil {
			l.ErrorContext(ctx, "gsgs get state", "err", err)
		} else if ok {
			columns = saved.Columns
		}

		// initial render
		grid, err := buildServiceGrid(
			ctx,
			studentReadModel,
			studentDTO.NewFilter(),
			iepReadModel,
			serviceReadModel,
			columns,
		)
		if err != nil {
			l.ErrorContext(ctx, "gsgs init", "err", err)
			return
		}
		sse.PatchElementTempl(pages.Grid(grid))

		for {
			select {
			case <-ctx.Done():
				return

			case <-notifier.Signal():
				grid, err := buildServiceGrid(
					ctx,
					studentReadModel,
					studentDTO.NewFilter(),
					iepReadModel,
					serviceReadModel,
					columns,
				)
				if err != nil {
					l.ErrorContext(ctx, "gsgs update", "err", err)
					continue
				}
				sse.PatchElementTempl(pages.Grid(grid))

			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.ServiceGrid{}
				if err := entry.JSON(signals); err != nil {
					l.ErrorContext(ctx, "gsgs json", "err", err)
					continue
				}
				if len(signals.Columns) == 0 {
					columns = sharedmodels.ServiceTypeList
				} else {
					columns = signals.Columns
				}
				grid, err := buildServiceGrid(
					ctx,
					studentReadModel,
					studentDTO.Filter{
						Search: signals.StudentFilter.Search,
					},
					iepReadModel,
					serviceReadModel,
					columns,
				)
				if err != nil {
					l.ErrorContext(ctx, "gsgs update", "err", err)
					continue
				}
				sse.PatchElementTempl(pages.Grid(grid))
			}
		}
	}
}

func queryServiceGrid(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		key := user.Username + ".servicegrid"

		signals := &dto.ServiceGrid{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qsg read", "err", err)
			return
		}

		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qsg put", "err", err)
		}
	}
}

// GET request to /services/create
// populates empty form with appropriate form type
func getServiceCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		props := dto.ServiceFormView{
			FormType: sharedmodels.FormTypeCreate,
		}
		_ = pages.Create(props).Render(ctx, w)
	}
}

// GET request to /services/create/stream
func getServiceCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// watches for view store changes
		key, err := getServiceViewstoreKey(sharedmodels.FormTypeCreate, user.Username, "")
		if err != nil {
			l.ErrorContext(ctx, "gscs vs key", "err", err, "username", user.Username)
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
			l.ErrorContext(ctx, "gscs watch", "err", err)
			return
		}
		defer watcher.Stop()

		// initialize view state to be updated by SSE
		if err := initializeServiceCreateState(ctx, l, vs, key); err != nil {
			l.ErrorContext(ctx, "gscs init vs", "err", err, "key", key)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				signals := &dto.ServiceFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.ErrorContext(ctx, "gscs json decode", "err", err)
					return
				}
				educators, _ := educatorReadModel.List(
					ctx,
					educatorEvents.FilterByRole(
						sharedmodels.EducatorRoleServiceProvider,
					),
				)
				view := dto.ServiceFormView{
					FormType:           sharedmodels.FormTypeCreate,
					Service:            signals.Service,
					ProviderSelectView: educatorDTO.NewSelectView(&signals.EducatorSelect.Filter, educators, signals.Service.ProviderID),
				}
				sse.PatchElementTempl(pages.Create(view))
			}
		}
	}
}

// QUERY request to /services/{formURL}/validate
func queryServiceFormValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &dto.ServiceFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "iep create validate signals read", "err", err)
			return
		}
		key, err := getServiceViewstoreKey(signals.FormType, user.Username, signals.Service.ID)
		if err != nil {
			l.ErrorContext(ctx, "qhfg vs key", "err", err, "form type", signals.FormType, "username", user.Username, "sid", signals.Service.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qhfg vs put state", "err", err, "key", key)
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
		signals := &dto.ServiceFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep services create signals", "err", err)
			return
		}
		if signals.Service.StudentID == "" {
			sse := newSSE(w, r)
			sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantError, "no student selected"))
			return
		}
		cmd := events.AddServiceToIEPCommand{
			Service:  signals.Service,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
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
		view := dto.NewServiceView(nil)
		_ = pages.View(view).Render(ctx, w)
	}
}

// GET request to /services/{id}/stream
func getServiceViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel *events.ReadModel,
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
		props := dto.ServiceFormView{
			FormType: sharedmodels.FormTypeEdit,
		}
		_ = pages.Edit(props).Render(ctx, w)
	}
}

// GET request to /service/{id}/stream
func getServiceEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	serviceReadModel *events.ReadModel,
	iepReadModel *iepEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
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
			l.ErrorContext(ctx, "gses sub", "err", err)
			return
		}
		defer sub.Close()

		// get view store key based on form type and info
		key, err := getServiceViewstoreKey(sharedmodels.FormTypeEdit, "", serviceID)
		if err != nil {
			l.ErrorContext(ctx, "gses vs key", "err", err, "id", serviceID)
			return
		}

		// watch for view store changes
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "gses vs watch", "err", err, "key", key)
			return
		}
		defer watcher.Stop()

		// check if view exists (someone is already editing the service)
		// if not, populate the view
		_, ok, err := vs.Get(ctx, key)
		if !ok {
			if err := refreshServiceEditState(
				ctx,
				l,
				vs,
				serviceID,
				serviceReadModel,
				iepReadModel,
			); err != nil {
				if err.Error() == "service not found" {
					sse.PatchElementTempl(pages.NotFound())
				}
				l.ErrorContext(ctx, "gses refresh state", "err", err)
				return
			}
		}
		if err != nil {
			l.ErrorContext(ctx, "gses get state", "err", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				if err := refreshServiceEditState(
					ctx,
					l,
					vs,
					serviceID,
					serviceReadModel,
					iepReadModel,
				); err != nil {
					if err.Error() == "service not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "gses notifier refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.ServiceFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.ErrorContext(ctx, "iep service edit stream json", "err", err)
					return
				}
				view := dto.ServiceFormView{
					FormType: sharedmodels.FormTypeEdit,
					Service:  signals.Service,
				}
				sse.PatchElementTempl(pages.Edit(view))
			}
		}
	}
}

// QUERY request to /services/{formURL}/{field}/{value}
func queryServiceFormField(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		field := chi.URLParam(r, "field")
		value := chi.URLParam(r, "value")
		signals := &dto.ServiceFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit validate signals", "err", err)
			return
		}
		switch field {
		case "servicetype":
			serviceType, err := sharedmodels.ParseServiceType(value)
			if err != nil {
				l.ErrorContext(ctx, "qsff st parse", "err", err)
				return
			}
			signals.Service.ServiceType = serviceType
		default:
			l.ErrorContext(ctx, "qpff", "invalid field in form", field)
			return
		}
		key, err := getServiceViewstoreKey(signals.FormType, user.Username, signals.Service.ID)
		if err != nil {
			l.ErrorContext(ctx, "qsff key", "err", err, "key", key)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qsff vs put state", "error", err)
			return
		}
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
		signals := &dto.ServiceFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit signals", "err", err)
			return
		}
		cmd := events.UpdateServiceCommand{
			Service:  signals.Service,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
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
		signals := &struct {
			Service dto.ServiceView `json:"service"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit signals", "err", err)
			return
		}
		_, err := events.DeleteServiceCommandHandler(ctx, events.DeleteServiceCommand{
			ServiceID: serviceID,
			IEPID:     signals.Service.IEPID,
			StudentID: signals.Service.StudentID,
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
func getServicesCSV(
	l *slog.Logger,
	serviceReadModel *events.ReadModel,
	iepReadModel *iepEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		diffs, err := loadServiceDiffs(ctx, iepReadModel, studentReadModel, serviceReadModel)
		if err != nil {
			l.ErrorContext(ctx, "pscsv load diffs", "err", err)
		}

		// render view
		view := dto.NewServiceDiffTableView(diffs)
		pages.ReadCSV(view).Render(ctx, w)
	}
}

// POST request to /services/csv
func postServicesCSV(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	serviceReadModel *events.ReadModel,
	iepReadModel *iepEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		diffs, err := loadServiceDiffs(ctx, iepReadModel, studentReadModel, serviceReadModel)
		if err != nil {
			l.ErrorContext(ctx, "pscsv load diffs", "err", err)
		}

		for _, diff := range diffs {
			if diff.Status == sharedmodels.DiffSame {
				continue
			}
			if diff.Status == sharedmodels.DiffNew {
				_, err := events.AddServiceToIEPCommandHandler(
					ctx,
					events.AddServiceToIEPCommand{
						Service: *diff.New,
					},
					saver,
					retriever,
				)
				if err != nil {
					l.ErrorContext(ctx, "ps csv add", "err", err)
					continue
				}
			}
			if diff.Status == sharedmodels.DiffUpdated {
				_, err := events.UpdateServiceCommandHandler(
					ctx,
					events.UpdateServiceCommand{
						Service: models.Service{
							ID:              diff.Old.ID,
							IEPID:           diff.Old.IEPID,
							StudentID:       diff.Old.StudentID,
							StudentMARSSID:  diff.Old.StudentMARSSID,
							ServiceName:     diff.New.ServiceName,
							ServiceType:     diff.New.ServiceType,
							IndirectMinutes: diff.New.IndirectMinutes,
							DirectMinutes:   diff.New.DirectMinutes,
							FrequencyCount:  diff.New.FrequencyCount,
							FrequencyType:   diff.New.FrequencyType,
							LocationID:      diff.New.LocationID,
							StartDate:       diff.New.StartDate,
							EndDate:         diff.New.EndDate,
							CreatedAt:       diff.Old.CreatedAt,
						},
					},
					saver,
					retriever,
				)
				if err != nil {
					l.ErrorContext(ctx, "ps csv update", "err", err)
					continue
				}
			}
			if diff.Status == sharedmodels.DiffAbsent {
				_, err := events.DeleteServiceCommandHandler(
					ctx,
					events.DeleteServiceCommand{
						ServiceID: diff.Old.ID,
						IEPID:     diff.Old.IEPID,
						StudentID: diff.Old.StudentID,
					},
					saver,
					retriever,
				)
				if err != nil {
					l.ErrorContext(ctx, "ps csv delete", "err", err)
					continue
				}
			}
		}

		sse := newSSE(w, r)
		sse.Redirect("/services")
	}
}

func initializeServiceCreateState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	key string,
) error {
	signals := dto.ServiceFormSignals{
		FormType: sharedmodels.FormTypeCreate,
		Service:  *models.NewService(),
		EducatorSelect: dto.EducatorSelectSignals{
			Filter: educatorDTO.NewFilter(),
		},
	}
	return viewstore.PutState(ctx, vs, key, signals)
}

func refreshServiceViewState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	serviceID string,
	serviceReadModel *events.ReadModel,
) error {
	model, err := serviceReadModel.Get(ctx, serviceID)
	if err != nil {
		return err
	}
	key := model.ID + ".view"
	return viewstore.PutState(ctx, vs, key, model)
}

func refreshServiceEditState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	serviceID string,
	serviceReadModel *events.ReadModel,
	iepReadModel *iepEvents.ReadModel,
) error {
	model, err := serviceReadModel.Get(ctx, serviceID)
	if err != nil {
		return err
	}
	if model == nil {
		return errors.New("model not found")
	}
	iep, _ := iepReadModel.Get(ctx, model.IEPID)
	model.StudentID = iep.StudentID
	key, err := getServiceViewstoreKey(sharedmodels.FormTypeEdit, "", model.ID)
	if err != nil {
		return err
	}
	signals := dto.ServiceFormSignals{
		FormType: sharedmodels.FormTypeEdit,
		Service:  *model,
		EducatorSelect: dto.EducatorSelectSignals{
			Filter: educatorDTO.NewFilter(),
		},
	}
	return viewstore.PutState(ctx, vs, key, signals)
}

func getServiceViewstoreKey(
	formType sharedmodels.FormType,
	username string,
	serviceID string,
) (string, error) {
	if formType == sharedmodels.FormTypeCreate && username != "" {
		return username + ".services.create", nil
	} else if formType == sharedmodels.FormTypeEdit && serviceID != "" {
		return serviceID + ".edit", nil
	} else {
		return "", errors.New("error getting viewstore key")
	}
}

func loadServiceDiffs(
	ctx context.Context,
	iepReadModel *iepEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
	serviceReadModel *events.ReadModel,
) ([]sharedmodels.Diff[models.Service], error) {
	file, err := os.OpenFile("iep_services.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer file.Close()

	csvServices := []*models.CSVService{}
	if err := gocsv.UnmarshalFile(file, &csvServices); err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}

	exclude := map[string]bool{
		"Shared paraprofessional":           true,
		"One-to-One paraprofessional (1-1)": true,
	}
	filtered := make([]*models.CSVService, 0, len(csvServices))
	for _, svc := range csvServices {
		if !exclude[svc.ServiceName] {
			filtered = append(filtered, svc)
		}
	}

	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimLeft(s, "0")
		return s
	}
	students, err := studentReadModel.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	marssMap := make(map[string]string, len(students))
	for _, student := range students {
		key := normalize(student.MARSSID)
		if key == "" || key == "0" {
			continue
		}
		marssMap[key] = student.ID
	}

	converted := make([]models.Service, 0)
	for _, csvSvc := range filtered {
		key := normalize(csvSvc.StudentMARSSID)
		if key == "" || key == "0" {
			continue
		}
		studentID, ok := marssMap[key]
		if !ok {
			continue
		}

		ieps, err := iepReadModel.ListForStudent(ctx, studentID)
		if err != nil || len(ieps) == 0 {
			continue
		}

		csvSvc.StudentID = studentID
		csvSvc.IEPID = ieps[0].ID
		converted = append(converted, models.NewModelFromCSVRow(*csvSvc))
	}

	dbServices, err := serviceReadModel.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list db services: %w", err)
	}
	for i, service := range dbServices {
		iep, err := iepReadModel.Get(ctx, service.IEPID)
		if err != nil {
			continue
		}
		dbServices[i].StudentID = iep.StudentID
	}

	return models.CompareServices(dbServices, converted), nil
}

// buildServiceGrid assembles the grid from the read models and the given visible columns.
// Only students with an IEP are included. Students with an IEP but no services still appear.
func buildServiceGrid(
	ctx context.Context,
	studentReadModel *studentEvents.ReadModel,
	studentFilter studentDTO.Filter,
	iepReadModel *iepEvents.ReadModel,
	serviceReadModel *events.ReadModel,
	columns []sharedmodels.ServiceType,
) (dto.ServiceGrid, error) {
	students, err := studentReadModel.List(
		ctx,
		studentEvents.WithSearchFilter(studentFilter.Search),
		studentEvents.WithSort("family_name", "ASC"),
	)
	if err != nil {
		return dto.ServiceGrid{}, err
	}

	servicesByStudent := make(map[string]map[sharedmodels.ServiceType][]models.Service, len(students))
	withIEP := make([]studentModels.Student, 0, len(students))

	for _, student := range students {
		ieps, err := iepReadModel.ListForStudent(ctx, student.ID)
		if err != nil || len(ieps) == 0 {
			continue
		}
		withIEP = append(withIEP, student)

		services, err := serviceReadModel.ListServicesForIEP(ctx, ieps[0].ID)
		if err != nil {
			continue
		}
		byType := make(map[sharedmodels.ServiceType][]models.Service, len(services))
		for _, s := range services {
			byType[s.ServiceType] = append(byType[s.ServiceType], s)
		}
		servicesByStudent[student.ID] = byType
	}

	return dto.NewServiceGrid(withIEP, servicesByStudent, columns), nil
}
