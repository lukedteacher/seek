package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	pse "seek/internal/features/periods_schedules/events"
	"seek/internal/features/schedules/blocks"
	"seek/internal/features/schedules/events"
	"seek/internal/features/schedules/pages"
	"seek/internal/views/dto"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedules/list", s.getSchedulesList)
	r.Get("/schedules/create", s.getScheduleCreate)
	r.Post("/schedules/create/validate", s.postScheduleCreateValidate)
	r.Post("/schedules/create", s.postScheduleCreate)
	r.Get("/schedules/{id}", s.getScheduleView)
	r.Get("/schedules/{id}/stream", s.getScheduleViewStream)
	r.Get("/schedules/{id}/periods/{pid}", s.getPeriodScheduleView)
	r.Get("/schedules/{id}/edit", s.getScheduleEdit)
	r.Get("/schedules/{id}/edit/stream", s.getScheduleEditStream)
	r.Post("/schedules/{id}/edit/validate", s.postScheduleEditValidate)
	r.Post("/schedules/{id}/edit", s.postScheduleEdit)
	r.Post("/schedules/{id}/delete", s.deleteSchedule)
}

// GET request to /schedules
func (s Server) getSchedulesList(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	schedules, err := s.Schedules.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.List(user, schedules).Render(r.Context(), w)
}

// GET request to /schedules/create
func (s Server) getScheduleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	emptySchedule := models.Schedule{}
	validation := events.Validate(emptySchedule)

	teachers, _ := s.Teachers.List(ctx)
	periods, _ := s.Periods.List(ctx)
	_ = pages.Create(user, emptySchedule, teachers, periods, validation, nil).Render(ctx, w)
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
	ctx := r.Context()
	user := currentUser(r)
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := events.CreateScheduleCommandHandler(ctx, events.CreateScheduleCommand{
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
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
func (s Server) getScheduleView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view, _ := s.newScheduleComponentViewModel(ctx, scheduleRes)
	_ = pages.View(user, view).Render(ctx, w)
}

// GET request to /schedules/{id}/periods/{pid} ??
func (s Server) getPeriodScheduleView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	periodID := chi.URLParam(r, "pid")
	periodRes, _ := s.Periods.Get(ctx, periodID)
	view, _ := s.newScheduleComponentViewModel(ctx, scheduleRes)
	pview, _ := dto.NewViewFromPeriod(periodRes)
	studentIDs, _ := s.PeriodsStudents.ListStudentIDsForPeriod(ctx, pview.ID)
	for i := range studentIDs {
		student, _ := s.Students.Get(ctx, studentIDs[i])
		studentView := dto.NewStudentViewFromModel(student)
		pview.Students = append(pview.Students, *studentView)
	}
	_ = pages.ViewWithPeriod(user, view, pview).Render(ctx, w)
}

// GET request to /schedules/{id}/stream
func (s Server) getScheduleViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(scheduleID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		println(err.Error())
		return
	}
	defer sub.Close()

	if err := s.refreshScheduleViewState(ctx, scheduleID); err != nil {
		println("svs first refresh: ", err.Error())
		return
	}

	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		scheduleID+".view",
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
			if err := s.refreshScheduleViewState(ctx, scheduleID); err != nil {
				println("svs second refresh: ", err.Error())
				if err.Error() == "schedule not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			var model models.Schedule
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			view, _ := s.newScheduleComponentViewModel(ctx, &model)
			sse.PatchElementTempl(pages.View(user, view))
		}
	}
}

// GET request to /schedules/{id}
func (s Server) getScheduleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	efvm, _ := s.newEditScheduleViewModel(ctx, scheduleRes)
	scvm, _ := s.newScheduleComponentViewModel(ctx, scheduleRes)
	_ = pages.Edit(user, efvm, scvm).Render(ctx, w)
}

// GET request to /schedules/{id}/edit/stream
func (s Server) getScheduleEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(scheduleID), func(context.Context, []byte) {
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
		scheduleID+".edit",
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
			if err := s.refreshScheduleEditState(ctx, scheduleID); err != nil {
				println(err.Error())
				if err.Error() == "schedule not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var model models.Schedule
			if err := entry.JSON(&model); err != nil {
				println(err.Error())
				return
			}
			efvm, _ := s.newEditScheduleViewModel(ctx, &model)
			scvm, _ := s.newScheduleComponentViewModel(ctx, &model)
			sse.PatchElementTempl(pages.Edit(user, efvm, scvm))
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
		ID:        scheduleID,
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
	}

	viewstore.PutState(ctx, s.ViewStore, scheduleID, model)
}

// POST request to /schedules/{id}/edit
func (s Server) postScheduleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
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
					Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
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
					Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
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
		Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST request to /schedules/{id}/delete
func (s Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	scheduleID := chi.URLParam(r, "id")
	_, err := events.DeleteScheduleCommandHandler(ctx, events.DeleteScheduleCommand{
		ScheduleID: scheduleID,
		Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}

// VIEW STATE HELPER FUNCTIONS
func (s Server) refreshScheduleViewState(ctx context.Context, scheduleID string) error {
	schedule, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		return err
	}

	return viewstore.PutState(ctx, s.ViewStore, schedule.ID+".view", schedule)
}

func (s Server) refreshScheduleEditState(ctx context.Context, scheduleID string) error {
	schedule, err := s.Schedules.Get(ctx, scheduleID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, schedule.ID+".edit", schedule)
}

// builds a view model for the edit schedule view
// TODO genericize it to be used for other contexts?
// TODO save it in the state?
func (s Server) newEditScheduleViewModel(ctx context.Context, sm *models.Schedule) (blocks.EditScheduleViewModel, error) {
	if sm == nil {
		return blocks.EditScheduleViewModel{}, nil
	}
	periodIDs, err := s.PeriodsSchedules.ListPeriodIDsForSchedule(ctx, sm.ID)
	if err != nil {
		return blocks.EditScheduleViewModel{}, fmt.Errorf("list schedule periods %s: %w", sm.ID, err)
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
		ID:        sm.ID,
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

func (s Server) newScheduleComponentViewModel(ctx context.Context, sm *models.Schedule) (dto.ScheduleView, error) {
	if sm == nil {
		return dto.ScheduleView{}, nil
	}
	periodIDs, err := s.PeriodsSchedules.ListPeriodIDsForSchedule(ctx, sm.ID)
	if err != nil {
		return dto.ScheduleView{}, fmt.Errorf("list period IDs: %w", err)
	}
	pcvs := make([]dto.PeriodView, 0, len(periodIDs))
	for _, periodID := range periodIDs {
		period, err := s.Periods.Get(ctx, periodID)
		if err != nil {
			println(err.Error())
		}
		view, err := dto.NewViewFromPeriod(period)
		if err != nil {
			println("new view error in new schedule component view model function: ", err.Error())
		}
		pcvs = append(pcvs, view)
	}
	return dto.ScheduleView{
		ID:      sm.ID,
		Title:   sm.Title,
		Periods: pcvs,
	}, nil
}

func (s Server) teacherToSignals(teacher *models.Teacher) models.TeacherSignals {
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
