package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/blocks"
	pdto "seek/internal/features/periods/dto"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	"seek/internal/features/periods/pages"
	psevents "seek/internal/features/periods_students/events"
	"seek/internal/views/dto"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get("/periods/list", s.getPeriodsList)
	r.Get("/periods/list/stream", s.getPeriodsListStream)
	r.Get("/periods/create", s.getPeriodCreate)
	r.Get("/periods/create/stream", s.getPeriodCreateStream)
	r.Post("/periods/create/validate", s.postPeriodCreateValidate)
	r.Post("/periods/create", s.postPeriodCreate)
	r.Get("/periods/{id}/view", s.getPeriodView)
	r.Get("/periods/{id}/view/stream", s.getPeriodViewStream)
	r.Get("/periods/{id}/edit", s.getPeriodEdit)
	r.Get("/periods/{id}/edit/stream", s.getPeriodEditStream)
	r.Post("/periods/{id}/edit/validate", s.postPeriodEditValidate)
	r.Post("/periods/{id}/edit", s.postPeriodEdit)
	r.Delete("/periods/{id}/delete", s.deletePeriod)
}

// GET request to /periods
func (s Server) getPeriodsList(w http.ResponseWriter, r *http.Request) {
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
	periods, err := s.Periods.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := pdto.NewPeriodTableView(periods)

	_ = pages.List(user, signals.View, view).Render(ctx, w)
}

// GET request to /periods/stream
func (s Server) getPeriodsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any periods
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
			periods, err := s.Periods.List(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			view := pdto.NewPeriodTableView(periods)

			sse.PatchElementTempl(pages.List(user, 0, view))
		}
	}
}

// GET request to /periods/create
// TODO figure out if there's a way to have this use the same form as edit?
func (s Server) getPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	empty := models.NewPeriod()
	students, err := s.Students.List(ctx)
	if err != nil {
		println(err.Error())
	}
	selected := []string{}
	view := blocks.NewPeriodCreateFormView(empty, students, selected)
	_ = pages.Create(user, view).Render(ctx, w)
}

// GET request to /periods/create/stream
func (s Server) getPeriodCreateStream(w http.ResponseWriter, r *http.Request) {
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
			var model models.Period
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			students, err := s.Students.List(ctx)
			if err != nil {
				println(err.Error())
			}
			view := blocks.NewPeriodCreateFormView(&model, students, model.StudentIDs)
			sse.PatchElementTempl(pages.Create(user, view))
		}
	}
}

// POST request to /periods/create/validate
func (s Server) postPeriodCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pcv signal read: ", err.Error())
		return
	}
	model := dto.NewPeriodFromFormView(&signals.Period)
	// saves the state to a view store so that the SSE can update
	// TODO look into a better name for the channel
	if err := viewstore.PutState(ctx, s.ViewStore, "new", model); err != nil {
		println("view store error ", err.Error())
	}
}

// POST request to /periods/create
func (s Server) postPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pc signal read: ", err.Error())
		return
	}
	model := dto.NewPeriodFromFormView(&signals.Period)
	validation := events.Validate(&model)
	if validation == nil {
		println("some error")
		return
		// TODO actually validate
	}
	createPeriodCommand := events.CreatePeriodCommand{
		Title:     model.Title,
		StartTime: model.StartTime,
		Duration:  model.Duration,
		Days:      model.Days,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.CreatePeriodCommandHandler(ctx, createPeriodCommand, s.EventSaver)
	if err != nil {
		println("ph cpch error: ", err.Error())
		return
	}
	studentIDs := strings.Split(signals.Period.StudentIDs, ",")
	for _, studentID := range studentIDs {
		periodStudentAddCommand := psevents.PeriodStudentAddCommand{
			PeriodID:  result.PeriodID,
			StudentID: studentID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		_, err := psevents.PeriodStudentAddCommandHandler(
			ctx,
			periodStudentAddCommand,
			s.EventSaver,
			s.EventRetriever,
		)
		if err != nil {
			println(err.Error())
		}
	}
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&dto.PeriodView{}, sse)
	})
}

// GET request to /periods/{id}
func (s Server) getPeriodView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	model, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}
	view, err := dto.NewViewFromPeriod(model)
	if err != nil {
		println("error: ", err.Error())
		return
	}
	studentIDs, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
	for i := range studentIDs {
		student, _ := s.Students.Get(ctx, studentIDs[i])
		studentView := dto.NewStudentViewFromModel(*student)
		view.Students = append(view.Students, *studentView)
	}
	_ = pages.View(user, view).Render(ctx, w)
}

// GET request to /periods/{id}/stream
func (s Server) getPeriodViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
		return
	}
	defer sub.Close()

	if err := s.refreshPeriodViewState(ctx, periodID); err != nil {
		println("pvs first refresh: ", err.Error())
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		periodID+".view",
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
			if err := s.refreshPeriodViewState(ctx, periodID); err != nil {
				println("pvs second refresh: ", err.Error())
				if err.Error() == "period not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			var model models.Period
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view, err := dto.NewViewFromPeriod(&model)
			if err != nil {
				println("error: ", err.Error())
			}
			studentIDs, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
			for i := range studentIDs {
				student, _ := s.Students.Get(ctx, studentIDs[i])
				studentView := dto.NewStudentViewFromModel(*student)
				view.Students = append(view.Students, *studentView)
			}
			sse.PatchElementTempl(pages.View(user, view))
		}
	}
}

// GET request to /periods/{id}/edit
func (s Server) getPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	model, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	all, err := s.Students.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	selected, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
	view := blocks.NewPeriodEditFormView(model, all, selected)
	_ = pages.Edit(user, view).Render(ctx, w)
}

// GET request to /period/{id}/stream
func (s Server) getPeriodEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
		return
	}
	defer sub.Close()

	// watches the period edit view state kv
	watcher, err := s.ViewStore.Watch(
		ctx,
		periodID+".edit",
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
			if err := s.refreshPeriodEditState(ctx, periodID); err != nil {
				println(err.Error())
				if err.Error() == "period not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model models.Period
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			all, err := s.Students.List(ctx)
			if err != nil {
				println(err.Error())
			}
			selected, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
			view := blocks.NewPeriodEditFormView(&model, all, selected)
			sse.PatchElementTempl(pages.Edit(user, view))
		}
	}
}

// POST request to /periods/{id}/edit/validate
func (s Server) postPeriodEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vep signals: ", err.Error())
		return
	}
	signals.Period.ID = periodID
	model := dto.NewPeriodFromFormView(&signals.Period)
	viewstore.PutState(ctx, s.ViewStore, periodID, model)
}

// POST request to /periods/{id}/edit
func (s Server) postPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("ep signals: ", err.Error())
		return
	}
	periodID := chi.URLParam(r, "id")
	updatePeriodCommand := events.UpdatePeriodCommand{
		Id:        periodID,
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      sharedmodels.DaysSignalsToDaysBitmask(signals.Period.Days),
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.UpdatePeriodCommandHandler(ctx, updatePeriodCommand, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(fmt.Errorf("upch error: %w", err))
		return
	}
	if result.Skipped == true {
		println("period update skipped")
	}
	current, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, periodID)
	proposed := strings.Split(signals.Period.StudentIDs, ",")
	println("p: ", len(proposed))
	if len(current) != 0 || len(proposed) != 0 {
		// build maps for O(1) lookups
		currentMap := make(map[string]bool)
		for _, v := range current {
			currentMap[v] = true
		}

		proposedMap := make(map[string]bool)
		for _, v := range proposed {
			proposedMap[v] = true
		}

		// find deletions
		for _, studentID := range current {
			if !proposedMap[studentID] {
				result, err := psevents.PeriodStudentRemoveCommandHandler(ctx, psevents.PeriodStudentRemoveCommand{
					PeriodID:  periodID,
					StudentID: studentID,
					Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
				}, s.EventSaver, s.EventRetriever)
				if result != nil {
					println("reid: ", result.EventID)
				}
				if err != nil {
					println("re: ", err.Error())
				}
			}
		}

		// find additions
		for _, studentID := range proposed {
			if !currentMap[studentID] {
				result, err := psevents.PeriodStudentAddCommandHandler(ctx, psevents.PeriodStudentAddCommand{
					PeriodID:  periodID,
					StudentID: studentID,
					Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
				}, s.EventSaver, s.EventRetriever)
				if result != nil {
					println("aeid: ", result.EventID)
				}
				if err != nil {
					println("ae: ", err.Error())
				}
			}
		}
	}
}

// DELETE request to /periods/{id}
func (s Server) deletePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	_, err := events.DeletePeriodCommandHandler(ctx, events.DeletePeriodCommand{
		PeriodID: periodID,
		Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(err.Error())
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/periods/list")
}

func (s Server) refreshPeriodViewState(ctx context.Context, periodID string) error {
	period, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, period.ID+".view", period)
}

func (s Server) refreshPeriodEditState(ctx context.Context, periodID string) error {
	period, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, period.ID+".edit", period)
}

func (s Server) periodToSignals(period *models.Period) models.PeriodSignals {
	if period == nil {
		return models.PeriodSignals{}
	}
	return models.PeriodSignals{
		ID:        period.ID,
		Title:     period.Title,
		StartTime: period.StartTime,
		Duration:  int(period.Duration),
		Days:      sharedmodels.DaysBitmaskToDaysSignals(period.Days),
	}
}

func (s Server) periodSignalsToModel(period *models.PeriodSignals) models.Period {
	if period == nil {
		return models.Period{}
	}
	return models.Period{
		ID:        period.ID,
		Title:     period.Title,
		StartTime: period.StartTime,
		Duration:  int64(period.Duration),
		Days:      sharedmodels.DaysSignalsToDaysBitmask(period.Days),
	}
}
