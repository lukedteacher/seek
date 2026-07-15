package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/schedules/blocks"
	"seek/internal/features/schedules/events"
	"seek/internal/features/schedules/pages"
	pse "seek/internal/features/periods_schedules/events"
	"seek/internal/views/dto"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedules", s.getSchedules)
	r.Get("/schedules/create", s.getScheduleCreate)
	r.Post("/schedules/create/validate", s.postScheduleCreateValidate)
	r.Post("/schedules/create", s.postScheduleCreate)
	r.Get("/schedules/{id}", s.getScheduleEdit)
	r.Get("/schedules/{id}/edit", s.getScheduleEdit)
	r.Get("/schedules/{id}/edit/stream", s.getScheduleEditStream)
	r.Post("/schedules/{id}/edit/validate", s.postScheduleEditValidate)
	r.Post("/schedules/{id}/edit", s.postScheduleEdit)
	r.Post("/schedules/{id}/delete", s.deleteSchedule)
}

// GET request to /schedules
func (s Server) getSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := s.Schedules.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.List(schedules).Render(r.Context(), w)
}

// GET request to /schedules/create
func (s Server) getScheduleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	emptySchedule := models.Schedule{}
	validation := events.Validate(emptySchedule)

	teachers, _ := s.Teachers.List(ctx)
	periods, _ := s.Periods.List(ctx)
	_ = pages.Create(emptySchedule, teachers, periods, validation, nil).Render(ctx, w)
}

// POST request to /schedules/create/validate
func (s Server) postScheduleCreateValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := models.Schedule{
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
	}

	selectedTeacher := &models.Teacher{}

	teachers, _ := s.Teachers.List(ctx)
	for _, teacher := range teachers {
		if teacher.ID == signals.Schedule.TeacherID {
			selectedTeacher = &teacher
		}
	}
	periods, _ := s.Periods.List(ctx)
	validation := events.Validate(model)
	patchTempl(w, r, blocks.CreateForm(model, teachers, periods, validation, selectedTeacher))
}

// POST request to /schedules/create
func (s Server) postScheduleCreate(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := events.CreateScheduleCommandHandler(r.Context(), events.CreateScheduleCommand{
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&Signals{}, sse)
	})
}

// GET request to /schedules/{id}
func (s Server) getScheduleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	efvm, _ := s.newEditScheduleViewModel(ctx, scheduleRes)
	scvm, _ := s.newScheduleComponentViewModel(ctx, scheduleRes)
	_ = pages.Edit(efvm, scvm).Render(ctx, w)
}

// GET request to /schedules/{id}/stream
func (s Server) getScheduleEditStream(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	ctx := r.Context()
	scheduleID := chi.URLParam(r, "id")

	updates := make(chan struct{}, 1)
	notify := func() {
		select {
		case updates <- struct{}{}:
		default:
		}
	}

	// watches that initial store
	watcher, err := s.ViewStore.Watch(
		ctx,
		scheduleID,
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		println(err.Error())
		return
	}
	defer watcher.Stop()

	sub, err := s.Subscriber.Subscribe(ctx, events.Channel("idk"), func(context.Context, []byte) {
		notify()
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
		case <-updates:
			if err := s.refreshScheduleViewState(ctx, scheduleID); err != nil {
				println(err.Error())
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model models.Schedule
			if err := entry.JSON(&model); err != nil {
				println("up: ", err.Error())
				return
			}
			efvm, _ := s.newEditScheduleViewModel(ctx, &model)
			scvm, _ := s.newScheduleComponentViewModel(ctx, &model)
			sse.PatchElementTempl(pages.Edit(efvm, scvm))
		}
	}
}

// POST request to /schedules/{id}/edit/validate
// saves the current schedule view
// will be validated through the SSE
func (s Server) postScheduleEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scheduleID := chi.URLParam(r, "id")
	model := models.Schedule{
		Id:        scheduleID,
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
	}

	viewstore.PutState(ctx, s.ViewStore, scheduleID, model)
}

// POST request to /schedules/{id}/edit
func (s Server) postScheduleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scheduleID := chi.URLParam(r, "id")
	currentPeriodIDs, _ := s.PeriodsSchedules.ListPeriodIDsForSchedule(ctx, scheduleID)
	proposedPeriodIDs := []string{}
	for _, periodID := range signals.Schedule.PeriodIDs {
		proposedPeriodIDs = append(proposedPeriodIDs, periodID)
	}
	if len(currentPeriodIDs) != 0 || len(proposedPeriodIDs) != 0 {
		// build maps for O(1) lookups
		currentMap := make(map[string]bool)
		for _, v := range currentPeriodIDs {
			currentMap[v] = true
		}

		proposedMap := make(map[string]bool)
		for _, v := range proposedPeriodIDs {
			proposedMap[v] = true
		}

		// find deletions
		for _, periodID := range currentPeriodIDs {
			if !proposedMap[periodID] {
				result, err := pse.PeriodScheduleRemoveCommandHandler(ctx, pse.PeriodScheduleRemoveCommand{
					ScheduleID: scheduleID,
					PeriodID:   periodID,
					Metadata:   eventstore.HTTPCommandMetadata(r),
				}, s.EventSaver, s.EventRetriever)
				if err != nil {
					println("re: ", err.Error())
				}
				if result != nil {
					println("reid: ", result.EventID)
				}
			}
		}

		// find additions
		for _, periodID := range proposedPeriodIDs {
			if !currentMap[periodID] {
				result, err := pse.PeriodScheduleAddCommandHandler(ctx, pse.PeriodScheduleAddCommand{
					ScheduleID: scheduleID,
					PeriodID:   periodID,
					Metadata:   eventstore.HTTPCommandMetadata(r),
				}, s.EventSaver, s.EventRetriever)
				if err != nil {
					println("ae: ", err.Error())
				}
				if result != nil {
					println("aeid: ", result.EventID)
				}
			}
		}
	}
	_, err := events.UpdateScheduleCommandHandler(ctx, events.UpdateScheduleCommand{
		Id:        scheduleID,
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST request to /schedules/{id}/delete
func (s Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "id")
	_, err := events.DeleteScheduleCommandHandler(r.Context(), events.DeleteScheduleCommand{
		ScheduleID: scheduleID,
		Metadata:   eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}

func (s Server) refreshScheduleViewState(ctx context.Context, scheduleID string) error {
	schedule, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		return err
	}

	return viewstore.PutState(ctx, s.ViewStore, scheduleID, schedule)
}

// builds a view model for the edit schedule view
// TODO genericize it to be used for other contexts?
// TODO save it in the state?
func (s *Server) newEditScheduleViewModel(ctx context.Context, sm *models.Schedule) (blocks.EditScheduleViewModel, error) {
	if sm == nil {
		return blocks.EditScheduleViewModel{}, nil
	}
	periodIDs, err := s.PeriodsSchedules.ListPeriodIDsForSchedule(ctx, sm.Id)
	if err != nil {
		return blocks.EditScheduleViewModel{}, fmt.Errorf("list schedule periods %s: %w", sm.Id, err)
	}

	schedulePeriodsSignals := make([]models.PeriodSignals, 0, len(periodIDs))
	for _, periodID := range periodIDs {
		pm, err := s.Periods.Get(ctx, periodID)
		if err != nil {
			return blocks.EditScheduleViewModel{}, fmt.Errorf("get period %s: %w", periodID, err)
		}
		periodSignals := s.periodToSignals(pm)
		schedulePeriodsSignals = append(schedulePeriodsSignals, periodSignals)
	}
	scheduleSignals := models.ScheduleSignals{
		ID:        sm.Id,
		Title:     sm.Title,
		TeacherID: sm.TeacherId,
		PeriodIDs: periodIDs,
	}

	validation := events.Validate(*sm)

	teachers, err := s.Teachers.List(ctx)
	if err != nil {
		return blocks.EditScheduleViewModel{}, fmt.Errorf("list teachers: %w", err)
	}
	teachersSignals := make([]models.TeacherSignals, 0, len(teachers))
	for i := range teachers {
		chosenName := ""
		if teachers[i].ChosenName != nil {
			chosenName = *teachers[i].ChosenName
		}
		teacherSignals := models.TeacherSignals{
			ID:         teachers[i].ID,
			FirstName:  teachers[i].FirstName,
			ChosenName: chosenName,
			LastName:   teachers[i].LastName,
		}
		teachersSignals = append(teachersSignals, teacherSignals)
	}

	periods, err := s.Periods.List(ctx)
	if err != nil {
		return blocks.EditScheduleViewModel{}, fmt.Errorf("list periods: %w", err)
	}
	periodsSignals := make([]models.PeriodSignals, 0, len(periods))
	for i := range periods {
		periodSignals := s.periodToSignals(&periods[i])
		periodsSignals = append(periodsSignals, periodSignals)
	}

	return blocks.EditScheduleViewModel{
		Schedule:        scheduleSignals,
		SchedulePeriods: schedulePeriodsSignals,
		Validation:      validation,
		Teachers:        teachersSignals,
		Periods:         periodsSignals,
	}, nil
}

func (s *Server) newScheduleComponentViewModel(ctx context.Context, sm *models.Schedule) (blocks.ScheduleComponentViewModel, error) {
	if sm == nil {
		return blocks.ScheduleComponentViewModel{}, nil
	}
	periodIDs, err := s.PeriodsSchedules.ListPeriodIDsForSchedule(ctx, sm.Id)
	if err != nil {
		return blocks.ScheduleComponentViewModel{}, fmt.Errorf("list period IDs: %w", err)
	}
	signals := models.ScheduleSignals{
		ID:        sm.Id,
		Title:     sm.Title,
		TeacherID: sm.TeacherId,
		PeriodIDs: periodIDs,
	}
	pcvms := make([]dto.PeriodView, 0, len(periodIDs))
	for _, periodID := range periodIDs {
		period, err := s.Periods.Get(ctx, periodID)
		if err != nil {
			println(err.Error())
		}
		view, err := dto.NewViewFromPeriod(period)
		if err != nil {
			println("error: ", err.Error())
		}
		pcvms = append(pcvms, view)
	}
	return blocks.ScheduleComponentViewModel{
		Schedule:         signals,
		PeriodViewModels: pcvms,
	}, nil
}

func (s *Server) teacherToSignals(teacher *models.Teacher) models.TeacherSignals {
	if teacher == nil {
		return models.TeacherSignals{}
	}
	chosenName := ""
	if teacher.ChosenName != nil {
		chosenName = *teacher.ChosenName
	}
	return models.TeacherSignals{
		ID:         teacher.ID,
		FirstName:  teacher.FirstName,
		ChosenName: chosenName,
		LastName:   teacher.LastName,
	}
}