package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	psevents "seek/internal/features/periods_students/events"
	"seek/internal/features/periods/blocks"
	"seek/internal/views/dto"
	pages "seek/internal/features/periods/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get("/periods", s.getPeriodsList)
	r.Get("/periods/create", s.getPeriodCreate)
	r.Get("/periods/create/stream", s.getPeriodCreateStream)
	r.Post("/periods/create/validate", s.postPeriodCreateValidate)
	r.Post("/periods/create", s.postPeriodCreate)
	r.Get("/periods/{id}", s.getPeriodView)
	r.Get("/periods/{id}/stream", s.getPeriodViewStream)
	r.Get("/periods/{id}/edit", s.getPeriodEdit)
	r.Get("/periods/{id}/edit/stream", s.getPeriodEditStream)
	r.Post("/periods/{id}/edit", s.postPeriodEdit)
	r.Post("/periods/{id}/edit/validate", s.postPeriodEditValidate)
	r.Delete("/periods/{id}", s.deletePeriod)
}

// GET request to /periods
func (s Server) getPeriodsList(w http.ResponseWriter, r *http.Request) {
	// defines the view (card or list or... something else?)
	type Signals struct {
		View int `json:"view"`
	}
	signals := &Signals{}
	datastar.ReadSignals(r, signals)

	periods, err := s.Periods.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.List(signals.View, periods).Render(r.Context(), w)
}

// GET request to /periods/create
// TODO figure out if there's a way to have this use the same form as edit?
func (s Server) getPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	empty := models.NewPeriod()
	students, err := s.Students.List(ctx)
	if err != nil {
		println(err.Error())
	}
	selected := []string{}
	view := blocks.NewPeriodCreateFormView(empty, students, selected)
	_ = pages.Create(view).Render(r.Context(), w)
}

// GET request to /periods/create/stream
func (s Server) getPeriodCreateStream(w http.ResponseWriter, r *http.Request) {
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
			sse.PatchElementTempl(pages.Create(view))
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
	if err := viewstore.PutState(ctx, s.ViewStore, "new", model); err != nil {
		println("view store error ", err.Error())
	}
}

// POST request to /periods/create
func (s Server) postPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Period dto.PeriodView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pc signal read: ", err.Error())
		return
	}
	model := dto.NewPeriodFromView(&signals.Period)
	validation := period.Validate(&model)
	if validation == nil {
		println("some error")
		return
		// TODO actually validate
	}
	createPeriodCommand := period.CreatePeriodCommand{
		Title:     model.Title,
		StartTime: model.StartTime,
		Duration:  model.Duration,
		Days:      model.Days,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}
	result, err := period.CreatePeriodCommandHandler(r.Context(), createPeriodCommand, s.EventSaver)
	if err != nil {
		println("ph cpch error: ", err.Error())
		return
	}

	for _, studentID := range signals.Period.StudentIDs {
		periodStudentAddCommand := psevents.PeriodStudentAddCommand{
			PeriodID:  result.PeriodID,
			StudentID: studentID,
			Metadata:  eventstore.HTTPCommandMetadata(r),
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
	periodID := chi.URLParam(r, "id")
	model, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound().Render(ctx, w)
	} else {
		view, err := dto.NewViewFromPeriod(model)
		if err != nil {
			println("error: ", err.Error())
		}
		studentIDs, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
		view.StudentIDs = studentIDs
		_ = pages.View(view).Render(ctx, w)
	}
}

// GET request to /periods/{id}/stream
func (s Server) getPeriodViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, period.Channel(periodID), func(context.Context, []byte) {
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
					sse.PatchElementTempl(pages.NotFound())
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
			view.StudentIDs = studentIDs
			sse.PatchElementTempl(pages.View(view))
		}
	}
}

// GET request to /periods/{id}/edit
func (s Server) getPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	model, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	all, err := s.Students.List(ctx)
	if err != nil {
		println(err.Error())
	}
	selected, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, model.ID)
	view := blocks.NewPeriodEditFormView(model, all, selected)
	_ = pages.Edit(view).Render(ctx, w)
}

// GET request to /period/{id}/stream
func (s Server) getPeriodEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, period.Channel(periodID), func(context.Context, []byte) {
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
					sse.PatchElementTempl(pages.NotFound())
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
			sse.PatchElementTempl(pages.Edit(view))
		}
	}
}

// POST request to /periods/{id}/edit/validate
func (s Server) postPeriodEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	signals := &struct {
		Period dto.PeriodView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vep signals: ", err.Error())
		return
	}
	signals.Period.ID = periodID
	model := dto.NewPeriodFromView(&signals.Period)
	viewstore.PutState(ctx, s.ViewStore, periodID, model)
}

// POST request to /periods/{id}/edit
func (s Server) postPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Period dto.PeriodView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("ep signals: ", err.Error())
		return
	}
	periodID := chi.URLParam(r, "id")
	days := models.DaysSignalsToDaysBitmask(signals.Period.Days)
	updatePeriodCommand := period.UpdatePeriodCommand{
		Id:        periodID,
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      days,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}
	result, err := period.UpdatePeriodCommandHandler(ctx, updatePeriodCommand, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(fmt.Errorf("upch error: %w", err))
		return
	}
	if result.Skipped == true {
		println("period update skipped")
	}
	current, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, periodID)
	proposed := signals.Period.StudentIDs
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
					Metadata:  eventstore.HTTPCommandMetadata(r),
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
					Metadata:  eventstore.HTTPCommandMetadata(r),
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
	periodID := chi.URLParam(r, "id")
	_, err := period.DeletePeriodCommandHandler(r.Context(), period.DeletePeriodCommand{
		PeriodID: periodID,
		Metadata: eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		println(err.Error())
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/periods")
}

func (s Server) refreshPeriodViewState(ctx context.Context, periodID string) error {
	period, err := s.Periods.Get(ctx, periodID)
	if err != nil {
		return err
	}
	if period.DeletedAt != "" {
		println("this period was deleted")
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

func (s *Server) periodToSignals(period *models.Period) models.PeriodSignals {
	if period == nil {
		return models.PeriodSignals{}
	}
	return models.PeriodSignals{
		ID:        period.ID,
		Title:     period.Title,
		StartTime: period.StartTime,
		Duration:  int(period.Duration),
		Days:      models.DaysBitmaskToDaysSignals(period.Days),
	}
}

func (s *Server) periodSignalsToModel(period *models.PeriodSignals) models.Period {
	if period == nil {
		return models.Period{}
	}
	return models.Period{
		ID:        period.ID,
		Title:     period.Title,
		StartTime: period.StartTime,
		Duration:  int64(period.Duration),
		Days:      models.DaysSignalsToDaysBitmask(period.Days),
	}
}
