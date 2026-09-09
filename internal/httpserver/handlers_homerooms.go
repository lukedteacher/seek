package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	educatorEvents "seek/internal/features/educators/events"
	"seek/internal/features/homerooms/dto"
	"seek/internal/features/homerooms/events"
	"seek/internal/features/homerooms/models"
	"seek/internal/features/homerooms/pages"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) homeroomRoutes(r chi.Router) {
	r.Get("/homerooms", getHomeroomsList(s.Logger))
	r.Get("/homerooms/stream", getHomeroomsListStream(s.Logger, s.Subscriber, s.ReadModels.Homerooms, s.ReadModels.Educators, s.ReadModels.Students))
	r.Get("/homerooms/create", getHomeroomCreate(s.Logger))
	r.Get("/homerooms/create/stream", getHomeroomCreateStream(s.Logger, s.ViewStore, s.ReadModels.Students, s.ReadModels.Educators))
	r.Post("/homerooms/create/validate", postHomeroomCreateValidate(s.Logger, s.ViewStore))
	r.Post("/homerooms/create/grades/{grade}", postHomeroomCreateGrades(s.Logger, s.ViewStore))
	r.Post("/homerooms/create/educators/{eid}", postHomeroomCreateEducators(s.Logger, s.ViewStore))
	r.Post("/homerooms/create/students/{sid}", postHomeroomCreateStudents(s.Logger, s.ViewStore))
	r.Post("/homerooms/create", postHomeroomCreate(s.Logger, s.EventSaver, s.EventRetriever))
	r.Get("/homerooms/{id}", getHomeroomView(s.Logger))
	r.Get("/homerooms/{id}/stream", getHomeroomViewStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Homerooms, s.ReadModels.Educators, s.ReadModels.Students))
	r.Get("/homerooms/{id}/edit", getHomeroomEdit(s.Logger))
	r.Get("/homerooms/{id}/edit/stream", getHomeroomEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Homerooms, s.ReadModels.Students, s.ReadModels.Educators))
	r.Post("/homerooms/{id}/edit/validate", postHomeroomEditValidate(s.Logger, s.ViewStore))
	r.Post("/homerooms/{id}/edit/grades/{grade}", postHomeroomEditGrades(s.Logger, s.ViewStore))
	r.Post("/homerooms/{id}/edit/educators/{eid}", postHomeroomEditEducators(s.Logger, s.ViewStore))
	r.Post("/homerooms/{id}/edit/students/{sid}", postHomeroomEditStudents(s.Logger, s.ViewStore))
	r.Post("/homerooms/{id}/edit", postHomeroomEdit(s.Logger, s.EventSaver, s.EventRetriever))
	r.Post("/homerooms/{id}/archive", postHomeroomArchive(s.Logger, s.EventSaver, s.EventRetriever))
	r.Delete("/homerooms/{id}", deleteHomeroom(s.Logger, s.EventSaver, s.EventRetriever))
}

// GET request to /homerooms
// renders an empty table template
// SSE will populate data
func getHomeroomsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewHomeroomTableView([]models.Homeroom{})
		_ = pages.List(view).Render(ctx, w)
	}
}

// GET request to /homerooms/stream
// populates data and keeps it updated with changes pushed from server
func getHomeroomsListStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	homeroomReadModel *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to any homerooms
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "hls subscribe", "err", err)
			return
		}
		defer sub.Close()
		homerooms, err := homeroomReadModel.ListWithIDs(ctx)
		if err != nil {
			l.ErrorContext(ctx, "hls list", "err", err)
		}
		view := dto.NewHomeroomTableView(homerooms)
		sse.PatchElementTempl(pages.List(view))

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				homerooms, err := homeroomReadModel.ListWithIDs(ctx)
				if err != nil {
					l.ErrorContext(ctx, "hls list", "err", err)
				}
				view := dto.NewHomeroomTableView(homerooms)
				sse.PatchElementTempl(pages.List(view))
			}
		}
	}
}

// GET request to /homerooms/create
func getHomeroomCreate(_ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.HomeroomFormView{}
		_ = pages.Create(view).Render(ctx, w)
	}
}

// GET request to /homerooms/create/stream
func getHomeroomCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
	studentReadModel *studentEvents.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// watch for view store changes
		key := user.Username + ".homerooms.create"
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "ghcs watch", "err", err)
			return
		}
		defer watcher.Stop()

		educators, _ := listEducators(ctx, l, educatorReadModel, nil)
		students := listStudents(ctx, l, studentReadModel, nil)
		view := dto.NewHomeroomFormView(
			&models.Homeroom{},
			students,
			nil,
			educators,
		)
		// since the model is a pointer, need this so signals are initalized properly
		view.EducatorIDs = []string{}
		view.StudentIDs = []string{}
		sse.PatchElementTempl(pages.Create(view))

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.Homeroom{}
				if err := entry.JSON(model); err != nil {
					l.Error("json decode", "err", err)
					return
				}
				educators, _ := listEducators(ctx, l, educatorReadModel, nil)
				students := listStudents(ctx, l, studentReadModel, nil)
				view := dto.NewHomeroomFormView(
					model,
					students,
					nil,
					educators,
				)
				sse.PatchElementTempl(pages.Create(view))
			}
		}
	}
}

// POST request to /homerooms/create/validate
// reads datastar signals from the form and saves them to the view store.
// this allows the SSE stream to detect changes and refresh the form preview.
func postHomeroomCreateValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "homeroom create validate signals", "err", err.Error())
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)

		// store the signals under a unique user key so the create stream can react
		key := user.Username + ".homerooms.create"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "post homeroom create validate viewstore", "err", err)
		}
	}
}

func postHomeroomCreateGrades(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		gradeString := chi.URLParam(r, "grade")
		gradeInt, err := strconv.Atoi(gradeString)
		grade := sharedmodels.Grade(gradeInt)
		if err != nil {
			l.ErrorContext(ctx, "phcg url", "err", err)
			return
		}
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phcg signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.GradesBitmask = *model.GradesBitmask.ToggleGrade(grade)
		key := user.Username + ".homerooms.create"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
			return
		}
	}
}

func postHomeroomCreateEducators(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		educatorID := chi.URLParam(r, "eid")
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phce signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.EducatorIDs = toggleID(model.EducatorIDs, educatorID)
		key := user.Username + ".homerooms.create"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "phce vs", "error", err)
		}
	}
}

func postHomeroomCreateStudents(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		studentID := chi.URLParam(r, "sid")
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phes signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.StudentIDs = toggleID(model.StudentIDs, studentID)
		key := user.Username + ".homerooms.create"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
		}
	}
}

// POST request to /homerooms/create
// reads the form signals, creates a new homeroom, syncs students and educators,
// then redirects to the homeroom view page.
func postHomeroomCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)

		signals := &struct {
			Homeroom dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post homeroom create signals", "err", err)
			return
		}

		// create the homeroom
		cmd := events.CreateHomeroomCommand{
			Homeroom: signals.Homeroom.Homeroom,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.CreateHomeroomCommandHandler(ctx, cmd, saver)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom create command handler", "err", err)
			return
		}

		homeroomID := result.EventID

		// sync educators (proposed list from form)
		secmd := events.SyncEducatorsInHomeroomCommand{
			HomeroomID:          homeroomID,
			ProposedEducatorIDs: signals.Homeroom.EducatorIDs,
		}
		if _, err := events.SyncEducatorsInHomeroomCommandHandler(ctx, secmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post homeroom create sync educators", "err", err)
		}

		// sync students (proposed list from form)
		sscmd := events.SyncStudentsInHomeroomCommand{
			HomeroomID:         homeroomID,
			ProposedStudentIDs: signals.Homeroom.StudentIDs,
		}
		if _, err := events.SyncStudentsInHomeroomCommandHandler(ctx, sscmd, saver, retriever); err != nil {
			l.ErrorContext(ctx, "post homeroom create sync students", "err", err)
		}

		// redirect to the new homeroom view
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/homerooms/%s", homeroomID))
	}
}

// GET request to /homerooms/{id}
func getHomeroomView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// minimal empty view
		view := dto.HomeroomView{}
		_ = pages.View(view).Render(ctx, w)
	}
}
func getHomeroomViewStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	homeroomReadModel *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		homeroomID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to the underlying model
		sub, err := subscriber.Subscribe(ctx, events.Channel(homeroomID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "get homeroom view stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// watches the key value stream for ephemeral changes
		// lasts 5m
		watcher, err := vs.Watch(
			ctx,
			homeroomID+".view",
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "get homeroom view stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshHomeroomViewState(ctx, l, homeroomID, homeroomReadModel, vs); err != nil {
			l.ErrorContext(ctx, "get homeroom view stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal(): // triggers when the read model publishes
				if err := refreshHomeroomViewState(ctx, l, homeroomID, homeroomReadModel, vs); err != nil {
					// the homeroom was deleted or archived
					if err.Error() == "homeroom not found" {
						sse.PatchElementTempl(pages.NotFound())
					}
					l.ErrorContext(ctx, "get homeroom view stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				homeroom := &models.Homeroom{}
				if err := entry.JSON(homeroom); err != nil {
					l.ErrorContext(ctx, "get homeroom view stream json", "err", err)
					return
				}
				view := dto.NewHomeroomView(homeroom)
				for i := range homeroom.EducatorIDs {
					educator, _ := educatorReadModel.GetByID(ctx, homeroom.EducatorIDs[i])
					educatorView := educatorDTO.NewEducatorView(educator)
					view.Educators = append(view.Educators, educatorView)
				}
				for i := range homeroom.StudentIDs {
					student, _ := studentReadModel.GetByID(ctx, homeroom.StudentIDs[i])
					studentView := studentDTO.NewView(student)
					view.Students = append(view.Students, studentView)
				}
				sse.PatchElementTempl(pages.View(view))
			}
		}
	}
}

func getHomeroomEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// creates a minimal view
		view := dto.HomeroomFormView{}
		_ = pages.Edit(view).Render(ctx, w)
	}
}

// GET request to /homerooms/{id}/edit/stream
// establishes an SSE connection that updates the edit form when the homeroom or view state changes.
func getHomeroomEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	homeroomsReadModel *events.ReadModel,
	studentReadModel *studentEvents.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		homeroomID := chi.URLParam(r, "id")
		sse := newSSE(w, r)

		// subscribe to event store changes
		notifier := NewDedupeNotifier()
		sub, err := subscriber.Subscribe(ctx, events.Channel(homeroomID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "homeroom edit stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		// check if the kv store has an edit view already created
		// aka someone else is editing the homeroom
		// if not, populate the view with data from the db
		key := homeroomID + ".edit"
		_, ok, err := vs.Get(ctx, key)
		if !ok {
			if err := refreshHomeroomEditState(
				ctx,
				l,
				homeroomID,
				homeroomsReadModel,
				vs,
			); err != nil {
				if err.Error() == "homeroom not found" {
					sse.PatchElementTempl(pages.NotFound())
				} else {
					l.ErrorContext(ctx, "refresh homeroom view state", "err", err)
				}
				return
			}
		}
		if err != nil {
			l.ErrorContext(ctx, "ghes get state", "vs get err", err)
		}

		// subscribe to the kv store for changes to the edit view state
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "ghes watch", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				// homeroom changed via event – refresh the view state and re‑render
				if err := refreshHomeroomEditState(
					ctx,
					l,
					homeroomID,
					homeroomsReadModel,
					vs,
				); err != nil {
					if err.Error() == "homeroom not found" {
						sse.PatchElementTempl(pages.NotFound())
					} else {
						l.ErrorContext(ctx, "refresh homeroom view state", "err", err)
					}
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				model := &models.Homeroom{}
				if err := entry.JSON(model); err != nil {
					l.Error("homeroom edit stream json", "err", err)
					return
				}
				educators, _ := listEducators(ctx, l, educatorReadModel, nil)
				students := listStudents(ctx, l, studentReadModel, nil)
				view := dto.NewHomeroomFormView(
					model,
					students,
					nil,
					educators,
				)
				sse.PatchElementTempl(pages.Edit(view))
			}
		}
	}
}

// POST request to /homerooms/{id}/edit/validate
// reads datastar signals from the form and saves them to the view store.
// this allows the SSE stream to detect changes and refresh the edit form preview.
func postHomeroomEditValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "homeroom edit validate signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		// store the signals under a key scoped to the homeroom
		key := model.ID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "post homeroom edit validate viewstore", "err", err)
		}
	}
}

func postHomeroomEditGrades(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		gradeString := chi.URLParam(r, "grade")
		gradeInt, err := strconv.Atoi(gradeString)
		grade := sharedmodels.Grade(gradeInt)
		if err != nil {
			l.ErrorContext(ctx, "pheg url", "err", err)
			return
		}
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phcg signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.GradesBitmask = *model.GradesBitmask.ToggleGrade(grade)
		key := model.ID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
			return
		}
	}
}

func postHomeroomEditEducators(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		educatorID := chi.URLParam(r, "eid")
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phee signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.EducatorIDs = toggleID(model.EducatorIDs, educatorID)
		key := model.ID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
		}
	}
}

func postHomeroomEditStudents(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		studentID := chi.URLParam(r, "sid")
		signals := &struct {
			FormView dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "phes signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromFormView(signals.FormView)
		model.StudentIDs = toggleID(model.StudentIDs, studentID)
		key := model.ID + ".edit"
		if err := viewstore.PutState(ctx, vs, key, model); err != nil {
			l.ErrorContext(ctx, "view store error", "error", err)
		}
	}
}

// POST request to /homerooms/{id}/edit
// reads the form signals, updates the homeroom, syncs educators and students,
// then redirects to the homeroom view page.
func postHomeroomEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)

		signals := &struct {
			Homeroom dto.HomeroomFormView `json:"homeroom"`
		}{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post homeroom edit signals", "err", err)
			return
		}

		homeroom := signals.Homeroom.Homeroom

		// update the homeroom itself
		cmd := events.UpdateHomeroomCommand{
			Homeroom: homeroom,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateHomeroomCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom edit command handler", "err", err, "pid", homeroom.ID)
			return
		}
		if result.Skipped {
			l.InfoContext(ctx, "post homeroom edit command handler", "skipped", result.Skipped)
		}

		// sync educators (proposed list from form)
		secmd := events.SyncEducatorsInHomeroomCommand{
			HomeroomID:          homeroom.ID,
			ProposedEducatorIDs: signals.Homeroom.EducatorIDs,
		}
		_, err = events.SyncEducatorsInHomeroomCommandHandler(ctx, secmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom create sync educators", "err", err)
		}

		// sync students (proposed list from form)
		sscmd := events.SyncStudentsInHomeroomCommand{
			HomeroomID:         homeroom.ID,
			ProposedStudentIDs: signals.Homeroom.StudentIDs,
		}
		_, err = events.SyncStudentsInHomeroomCommandHandler(ctx, sscmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom create sync students", "err", err)
		}

		// redirect to the homeroom view
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/homerooms/%s", homeroom.ID))
	}
}

// POST request to /homerooms/{id}/archive
func postHomeroomArchive(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		homeroomID := chi.URLParam(r, "id")
		_, err := events.ArchiveHomeroomCommandHandler(ctx, events.ArchiveHomeroomCommand{
			HomeroomID: homeroomID,
			Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "archive homeroom command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/homerooms")
	}
}

// DELETE request to /homerooms/{id}
func deleteHomeroom(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		homeroomID := chi.URLParam(r, "id")
		_, err := events.DeleteHomeroomCommandHandler(ctx, events.DeleteHomeroomCommand{
			HomeroomID: homeroomID,
			Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "delete homeroom command handler", "err", err)
			return
		}
		sse := newSSE(w, r)
		sse.Redirect("/homerooms")
	}
}

// gets homeroom from the db and saves it to a kv store
func refreshHomeroomViewState(
	ctx context.Context,
	l *slog.Logger,
	homeroomID string,
	homerooms *events.ReadModel,
	vs viewstore.Store,
) error {
	model, err := homerooms.GetWithIDs(ctx, homeroomID)
	if err != nil {
		return err
	}
	key := model.ID + ".view"
	return viewstore.PutState(ctx, vs, key, model)
}

// gets homeroom data from the db, converts it to a form view, and saves it to the store
func refreshHomeroomEditState(
	ctx context.Context,
	_ *slog.Logger,
	homeroomID string,
	homerooms *events.ReadModel,
	vs viewstore.Store,
) error {
	model, err := homerooms.GetWithIDs(ctx, homeroomID)
	if err != nil {
		return err
	}
	key := model.ID + ".edit"
	return viewstore.PutState(ctx, vs, key, model)
}

func toggleID(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			// remove it
			return append(slice[:i], slice[i+1:]...)
		}
	}
	// not found – add it
	return append(slice, value)
}
