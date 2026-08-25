package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/features/ieps/dto"
	"seek/internal/features/ieps/events"
	"seek/internal/features/ieps/models"
	"seek/internal/features/ieps/pages"
	studentEvents "seek/internal/features/students/events"
	"seek/internal/ui/core/coreblocks/toasts"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) iepRoutes(r chi.Router) {
	r.Get("/ieps", getIEPsList(s.Logger))
	r.Get("/ieps/stream", getIEPsListStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPs))
	r.Get("/ieps/create", getIEPCreate(s.Logger))
	r.Get("/ieps/create/stream", getIEPCreateStream(s.Logger, s.ViewStore, *s.ReadModels.Students))
	r.Post("/ieps/create/validate", postIEPCreateValidate(s.Logger, s.ViewStore))
	r.Post("/ieps/create", postIEPCreate(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/ieps/{id}", getIEPView(s.Logger))
	r.Get("/ieps/{id}/stream", getIEPViewStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPs))
	r.Get("/ieps/{id}/edit", getIEPEdit(s.Logger))
	r.Get("/ieps/{id}/edit/stream", getIEPEditStream(s.Logger, s.Subscriber, s.ViewStore, *s.ReadModels.IEPs, *s.ReadModels.Students))
	r.Post("/ieps/{id}/edit", postIEPEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/ieps/{id}/edit/validate", postIEPEditValidate(s.Logger, s.ViewStore))
	r.Delete("/ieps/{id}", deleteIEP(s.Logger, s.EventSaver, s.EventRetriever))
}

// GET request to /ieps
func getIEPsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewIEPTableView([]models.IEP{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /ieps/stream
func getIEPsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	_ viewstore.Store,
	iepReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any ieps
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep list stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		ieps, err := iepReadModel.List(ctx)
		l.Debug("test", "l", len(ieps))
		if err != nil {
			l.ErrorContext(ctx, "iep list stream db list", "err", err)
			return
		}
		view := dto.NewIEPTableView(ieps)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				// for now just refreshes the page
				// consider adding a view store for the list
				ieps, err := iepReadModel.List(ctx)
				l.Debug("test", "l", len(ieps))
				if err != nil {
					l.ErrorContext(ctx, "iep list stream db list", "err", err)
					return
				}
				view := dto.NewIEPTableView(ieps)
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /ieps/create
func getIEPCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Create(dto.IEPFormView{FormType: "create"}).Render(ctx, w)
	}
}

// GET request to /ieps/create/stream
func getIEPCreateStream(
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
		view := dto.NewIEPFormView(
			"create",
			&models.IEP{},
			students,
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
				model := &models.IEP{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep create stream watcher update", "err", err)
					return
				}
				students, _ := studentReadModel.List(ctx)
				view := dto.NewIEPFormView(
					"create",
					model,
					students,
				)
				sse.PatchElementTempl(pages.Create(view))
			}
		}
	}
}

// POST request to /ieps/create/validate
func postIEPCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			View dto.IEPView `json:"iep"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "iep create validate signals read", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		// saves the state to a view store so that the SSE can update
		// TODO look into a better name for the channel
		if err := viewstore.PutState(ctx, vs, "new", model); err != nil {
			l.ErrorContext(ctx, "post iep create validate viewstore", "err", err)
		}
	}
}

// POST request to /ieps/create
func postIEPCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			View dto.IEPView `json:"iep"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep create signals", "err", err)
			return
		}
		if signals.View.StudentID == "" {
			sse := newSSE(w, r)
			sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantError, "no student selected"))
			return
		}
		iep := events.IEPState{
			StudentID:   signals.View.StudentID,
			StartDate:   signals.View.StartDate.String(),
			EndDate:     signals.View.EndDate.String(),
			AmendedDate: signals.View.AmendedDate.String(),
		}
		cmd := events.AddIEPToStudentCommand{
			IEPState: iep,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.AddIEPToStudentCommandHandler(ctx, cmd, saver, retriever)
		sse := newSSE(w, r)
		if err != nil {
			l.ErrorContext(ctx, "post iep create command handler", "err", err)
			sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantError, err.Error()))
			return
		}
		sse.Redirect(fmt.Sprintf("/ieps/%s", result.EventID))
	}
}

// GET request to /ieps/{id}
func getIEPView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewIEPView(&models.IEP{})
		_ = pages.View(view).Render(ctx, w)
	}
}

// GET request to /ieps/{id}/stream
func getIEPViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	iepReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(iepID), func(context.Context, []byte) {
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
			iepID+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshIEPViewState(ctx, l, vs, iepID, iepReadModel); err != nil {
			l.ErrorContext(ctx, "iep service view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshIEPViewState(ctx, l, vs, iepID, iepReadModel); err != nil {
					if err.Error() == "iep not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				model := &models.IEP{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service view stream json", "err", err)
					return
				}
				view := dto.NewIEPView(model)
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

// GET request to /ieps/{id}/edit
func getIEPEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Edit(dto.IEPFormView{FormType: "edit"}).Render(ctx, w)
	}
}

// GET request to /iep/{id}/stream
func getIEPEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	iepReadModel events.ReadModel,
	studentReadModel studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(iepID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the iep edit view state kv
		watcher, err := vs.Watch(
			ctx,
			iepID+".edit",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "iep service edit stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshIEPEditState(ctx, l, vs, iepID, iepReadModel); err != nil {
			if err.Error() == "iep not found" {
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
				if err := refreshIEPEditState(ctx, l, vs, iepID, iepReadModel); err != nil {
					if err.Error() == "iep not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "iep service edit stream refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.IEP{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "iep service edit stream json", "err", err)
					return
				}
				students, _ := studentReadModel.List(ctx)
				view := dto.NewIEPFormView(
					"edit",
					model,
					students,
				)
				sse.PatchElementTempl(pages.Edit(view))
			}
		}
	}
}

// POST request to /ieps/{id}/edit/validate
func postIEPEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iepID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.IEPView `json:"iep"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit validate signals", "err", err)
			return
		}
		model := dto.NewModelFromView(&signals.View)
		viewstore.PutState(ctx, vs, iepID, model)
	}
}

// POST request to /ieps/{id}/edit
func postIEPEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		iepID := chi.URLParam(r, "id")
		signals := &struct {
			View dto.IEPView `json:"iep"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post iep service edit signals", "err", err)
			return
		}
		iep := events.IEPState{
			ID:          iepID,
			StudentID:   signals.View.StudentID,
			StartDate:   signals.View.StartDate.String(),
			EndDate:     signals.View.EndDate.String(),
			AmendedDate: signals.View.AmendedDate.String(),
		}
		command := events.UpdateIEPCommand{
			IEP:      iep,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateIEPCommandHandler(ctx, command, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post iep service edit command handler", "err", err)
			return
		}
		if result.Skipped == true {
			l.Info("post iep service edit command handler", "skipped", result.Skipped)
		}
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/ieps/%s", iepID))
	}
}

// DELETE request to /ieps/{id}
func deleteIEP(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		iepID := chi.URLParam(r, "id")
		_, err := events.DeleteIEPCommandHandler(ctx, events.DeleteIEPCommand{
			IEPID:    iepID,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete iep service command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/ieps")
	}
}

func refreshIEPViewState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	iepID string,
	iepReadModel events.ReadModel,
) error {
	iep, err := iepReadModel.Get(ctx, iepID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, iep.ID+".view", iep)
}

func refreshIEPEditState(
	ctx context.Context,
	_ *slog.Logger,
	vs viewstore.Store,
	iepID string,
	iepReadModel events.ReadModel,
) error {
	iep, err := iepReadModel.Get(ctx, iepID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, iep.ID+".edit", iep)
}
