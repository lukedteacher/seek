package httpserver

import (
	"context"
	"errors"
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
	// homeroom list
	r.Get("/homerooms", getHomeroomsList(s.Logger))
	r.Get("/homerooms/stream", getHomeroomsListStream(s.Logger, s.Subscriber, s.ReadModels.Homerooms, s.ReadModels.Educators, s.ReadModels.Students))
	// homeroom creation
	r.Get("/homerooms/create", getHomeroomCreate(s.Logger))
	r.Get("/homerooms/create/stream", getHomeroomCreateStream(s.Logger, s.ViewStore, s.ReadModels.Students, s.ReadModels.Educators))
	r.Query("/homerooms/create/validate", queryHomeroomFormValidate(s.Logger, s.ViewStore))
	r.Query("/homerooms/create/{field}/{value}", queryHomeroomFormField(s.Logger, s.ViewStore))
	r.Post("/homerooms/create", postHomeroomCreate(s.Logger, s.EventSaver, s.EventRetriever))
	// homeroom viewing
	r.Get("/homerooms/{id}", getHomeroomView(s.Logger))
	r.Get("/homerooms/{id}/stream", getHomeroomViewStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Homerooms, s.ReadModels.Educators, s.ReadModels.Students))
	// homeroom editing
	r.Get("/homerooms/{id}/edit", getHomeroomEdit(s.Logger))
	r.Get("/homerooms/{id}/edit/stream", getHomeroomEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Homerooms, s.ReadModels.Educators, s.ReadModels.Students))
	r.Query("/homerooms/{id}/edit/validate", queryHomeroomFormValidate(s.Logger, s.ViewStore))
	r.Query("/homerooms/{id}/edit/{field}/{value}", queryHomeroomFormField(s.Logger, s.ViewStore))
	r.Post("/homerooms/{id}/edit", postHomeroomEdit(s.Logger, s.ViewStore, s.EventSaver, s.EventRetriever))
	// other homeroom stuff
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
			l.ErrorContext(ctx, "hls sub", "err", err)
			return
		}
		defer sub.Close()

		// initialize data
		homerooms, err := homeroomReadModel.ListWithIDs(ctx)
		if err != nil {
			l.ErrorContext(ctx, "hls list", "err", err)
		}

		// make the view and push it via SSE
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
// populates empty form with appropriate form type
func getHomeroomCreate(
	l *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		props := pages.FormProps{
			FormType: sharedmodels.FormTypeCreate,
		}
		_ = pages.Form(props).Render(ctx, w)
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
		key, err := getHomeroomViewstoreKey(sharedmodels.FormTypeCreate, user.Username, "")
		if err != nil {
			l.ErrorContext(ctx, "ghcs vs key", "err", err, "username", user.Username)
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
			l.ErrorContext(ctx, "ghcs watch", "err", err)
			return
		}
		defer watcher.Stop()

		// initialize view state to be updated by SSE
		if err := initializeHomeroomCreateState(ctx, l, vs, key); err != nil {
			l.ErrorContext(ctx, "ghcs init vs", "err", err, "key", key)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.HomeroomFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.Error("ghcs json decode", "err", err)
					return
				}
				educators := listEducatorsByIDs(ctx, l, educatorReadModel, signals.Homeroom.EducatorIDs)
				students := listStudentsByIDs(ctx, l, studentReadModel, signals.Homeroom.StudentIDs)
				props := pages.FormProps{
					FormType:           sharedmodels.FormTypeCreate,
					Homeroom:           signals.Homeroom,
					EducatorSelectView: createEducatorSelectView(ctx, l, educatorReadModel, &signals.EducatorSelect.Filter, signals.Homeroom.EducatorIDs),
					SelectedEducators:  educatorDTO.NewViews(educators),
					StudentSelectView:  createStudentSelectView(ctx, l, studentReadModel, &signals.StudentSelect.Filter, signals.Homeroom.StudentIDs),
					SelectedStudents:   studentDTO.NewViews(students),
				}
				sse.PatchElementTempl(pages.Form(props))
			}
		}
	}
}

// QUERY request to /homerooms/create/validate or
// QUERY request to  /homerooms/{id}/edit/validate
// reads datastar signals from the form and saves them to the view store
// this allows the SSE stream to detect changes and refresh the form with validation
func queryHomeroomFormValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &dto.HomeroomFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qhfv read signals", "err", err)
			return
		}
		// store the signals in the viewstore based on form type
		key, err := getHomeroomViewstoreKey(signals.FormType, user.Username, signals.Homeroom.ID)
		if err != nil {
			l.ErrorContext(ctx, "qhfg vs key", "err", err, "form type", signals.FormType, "username", user.Username, "hid", signals.Homeroom.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qhfv vs put state", "err", err, "key", key)
		}
	}
}

// QUERY request to /homerooms/create/grades/{grade}
// toggles selected grade
func queryHomeroomFormField(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		field := chi.URLParam(r, "field")
		value := chi.URLParam(r, "value")
		signals := &dto.HomeroomFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qhff signals", "err", err)
			return
		}
		switch field {
		case "grades":
			gradeInt, err := strconv.Atoi(value)
			if err != nil {
				l.ErrorContext(ctx, "qhfg a to i", "err", err)
				return
			}
			grade := sharedmodels.Grade(gradeInt)
			signals.Homeroom.GradesBitmask.ToggleGrade(grade)
		case "educators":
			signals.Homeroom.EducatorIDs = toggleID(signals.Homeroom.EducatorIDs, value)
		case "students":
			signals.Homeroom.StudentIDs = toggleID(signals.Homeroom.StudentIDs, value)
		default:
			l.ErrorContext(ctx, "qhff", "invalid field in form", field)
			return
		}
		key, err := getHomeroomViewstoreKey(signals.FormType, user.Username, signals.Homeroom.ID)
		if err != nil {
			l.ErrorContext(ctx, "qhff vs key", "err", err, "form type", signals.FormType, "username", user.Username, "hid", signals.Homeroom.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qhff vs", "error", err)
			return
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
		signals := &dto.HomeroomFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post homeroom create signals", "err", err)
			return
		}

		// create the homeroom
		model := dto.NewHomeroomModelFromView(signals.Homeroom)
		cmd := events.CreateHomeroomCommand{
			Homeroom: model,
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
		view := pages.ViewProps{}
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

		// subscribes to the channel which publishes changes to the underlying model
		notifier := NewDedupeNotifier()
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
				model := &models.Homeroom{}
				if err := entry.JSON(model); err != nil {
					l.ErrorContext(ctx, "get homeroom view stream json", "err", err)
					return
				}
				view := pages.ViewProps{
					Homeroom: dto.NewHomeroomView(model),
				}
				for i := range view.Homeroom.EducatorIDs {
					educator, _ := educatorReadModel.GetByID(ctx, view.Homeroom.EducatorIDs[i])
					educatorView := educatorDTO.NewView(educator)
					view.Educators = append(view.Educators, educatorView)
				}
				for i := range view.Homeroom.StudentIDs {
					student, _ := studentReadModel.GetByID(ctx, view.Homeroom.StudentIDs[i])
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
		props := pages.FormProps{
			FormType: sharedmodels.FormTypeEdit,
		}
		_ = pages.Form(props).Render(ctx, w)
	}
}

// GET request to /homerooms/{id}/edit/stream
// establishes an SSE connection that updates the edit form when the homeroom or view state changes.
func getHomeroomEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	homeroomsReadModel *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
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
			l.ErrorContext(ctx, "ghes sub", "err", err)
			return
		}
		defer sub.Close()

		// get view store key based on form type and info
		key, err := getHomeroomViewstoreKey(sharedmodels.FormTypeEdit, "", homeroomID)
		if err != nil {
			l.ErrorContext(ctx, "ghes vs key", "err", err)
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
			l.ErrorContext(ctx, "ghes vs watch", "err", err, "key", key)
			return
		}
		defer watcher.Stop()

		// check if the kv store has an edit view already created
		// aka someone else is editing the homeroom
		// if not, populate the view with data from the db
		_, ok, err := vs.Get(ctx, key)
		if !ok {
			if err := refreshHomeroomEditState(
				ctx,
				l,
				homeroomID,
				homeroomsReadModel,
				educatorReadModel,
				studentReadModel,
				vs,
			); err != nil {
				if err.Error() == "homeroom not found" {
					sse.PatchElementTempl(pages.NotFound())
				} else {
					l.ErrorContext(ctx, "ghes refresh state", "err", err)
				}
				return
			}
		}
		if err != nil {
			l.ErrorContext(ctx, "ghes get state", "err", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				// if homeroom changed via event, refresh state (will trigger SSE re-render)
				if err := refreshHomeroomEditState(
					ctx,
					l,
					homeroomID,
					homeroomsReadModel,
					educatorReadModel,
					studentReadModel,
					vs,
				); err != nil {
					if err.Error() == "homeroom not found" {
						sse.PatchElementTempl(pages.NotFound())
					} else {
						l.ErrorContext(ctx, "ghes notifier refresh", "err", err)
					}
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				signals := &dto.HomeroomFormSignals{}
				if err := entry.JSON(signals); err != nil {
					l.Error("ghes watcher json", "err", err)
					return
				}
				educators := listEducatorsByIDs(ctx, l, educatorReadModel, signals.Homeroom.EducatorIDs)
				students := listStudentsByIDs(ctx, l, studentReadModel, signals.Homeroom.StudentIDs)
				props := pages.FormProps{
					FormType:           sharedmodels.FormTypeEdit,
					Homeroom:           signals.Homeroom,
					EducatorSelectView: createEducatorSelectView(ctx, l, educatorReadModel, &signals.EducatorSelect.Filter, signals.Homeroom.EducatorIDs),
					SelectedEducators:  educatorDTO.NewViews(educators),
					StudentSelectView:  createStudentSelectView(ctx, l, studentReadModel, &signals.StudentSelect.Filter, signals.Homeroom.StudentIDs),
					SelectedStudents:   studentDTO.NewViews(students),
				}
				sse.PatchElementTempl(pages.Form(props))
			}
		}
	}
}

// POST request to /homerooms/{id}/edit
// reads the form signals, updates the homeroom, syncs educators and students,
// then redirects to the homeroom view page.
func postHomeroomEdit(
	l *slog.Logger,
	vs viewstore.Store,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		homeroomID := chi.URLParam(r, "id")
		signals := &dto.HomeroomFormSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "post homeroom edit signals", "err", err)
			return
		}
		model := dto.NewHomeroomModelFromView(signals.Homeroom)
		// update the homeroom itself
		cmd := events.UpdateHomeroomCommand{
			Homeroom: model,
			Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateHomeroomCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom edit command handler", "err", err, "pid", model.ID)
			return
		}
		if result.Skipped {
			l.InfoContext(ctx, "post homeroom edit command handler", "skipped", result.Skipped)
		}

		// sync educators (proposed list from form)
		secmd := events.SyncEducatorsInHomeroomCommand{
			HomeroomID:          model.ID,
			ProposedEducatorIDs: model.EducatorIDs,
		}
		_, err = events.SyncEducatorsInHomeroomCommandHandler(ctx, secmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom edit sync educators", "err", err)
		}

		// sync students (proposed list from form)
		sscmd := events.SyncStudentsInHomeroomCommand{
			HomeroomID:         model.ID,
			ProposedStudentIDs: model.StudentIDs,
		}
		_, err = events.SyncStudentsInHomeroomCommandHandler(ctx, sscmd, saver, retriever)
		if err != nil {
			l.ErrorContext(ctx, "post homeroom edit sync students", "err", err)
		}
		key := homeroomID + ".edit"
		if err := vs.Delete(ctx, key); err != nil {
			l.ErrorContext(ctx, "phe delete viewstore state", "err", err)
		}
		// redirect to the homeroom view
		sse := newSSE(w, r)
		sse.Redirect(fmt.Sprintf("/homerooms/%s", model.ID))
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

func initializeHomeroomCreateState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	key string,
) error {
	signals := dto.NewHomeroomFormSignals(sharedmodels.FormTypeCreate, nil)
	return viewstore.PutState(ctx, vs, key, signals)
}

// gets homeroom data from the db, converts it to a form view, and saves it to the store
func refreshHomeroomEditState(
	ctx context.Context,
	l *slog.Logger,
	homeroomID string,
	homeroomEvents *events.ReadModel,
	educatorReadModel *educatorEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
	vs viewstore.Store,
) error {
	model, err := homeroomEvents.GetWithIDs(ctx, homeroomID)
	if err != nil {
		return err
	}
	studentGradeFilter := make(map[string]bool, len(sharedmodels.GradeList))
	if model.GradesBitmask != -1 {
		for _, grade := range sharedmodels.GradeList {
			if model.GradesBitmask.IsGradeSet(grade) {
				studentGradeFilter[grade.String()] = true
			} else {
				studentGradeFilter[grade.String()] = false
			}
		}
	}
	studentPlanFilter := make(map[string]bool, len(sharedmodels.PlanTypeList))
	for _, planType := range sharedmodels.PlanTypeList {
		studentPlanFilter[planType.String()] = true
	}
	educators, _ := listEducators(ctx, l, educatorReadModel, nil)
	studentFilter := &studentDTO.Filter{
		Grade:    studentGradeFilter,
		PlanType: studentPlanFilter,
	}
	students := listStudents(ctx, l, studentReadModel, studentFilter)
	view := dto.NewHomeroomFormView(
		model,
		students,
		studentFilter,
		educators,
	)
	key, err := getHomeroomViewstoreKey(sharedmodels.FormTypeEdit, "", homeroomID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, key, view)
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

func getHomeroomViewstoreKey(
	formType sharedmodels.FormType,
	username string,
	homeroomID string,
) (string, error) {
	if formType == sharedmodels.FormTypeCreate && username != "" {
		return username + ".homerooms.create", nil
	} else if formType == sharedmodels.FormTypeEdit && homeroomID != "" {
		return homeroomID + ".edit", nil
	} else {
		return "", errors.New("error getting viewstore key")
	}
}
