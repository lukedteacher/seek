package httpserver

import (
	"context"
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/schedule"
	"seek/internal/views/blocks"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedule", s.schedule)
	r.Get("/schedules/create", s.createScheduleForm)
	r.Post("/schedules/create/validate", s.validateCreateSchedule)
	r.Post("/schedules/create", s.createSchedule)
	r.Get("/schedules/{id}", s.getEditSchedule)
	r.Get("/schedules/{id}/edit", s.getEditSchedule)
	r.Get("/schedules/{id}/edit/stream", s.editScheduleStream)
	r.Post("/schedules/{id}/edit/validate", s.validateEditSchedule)
	r.Post("/schedules/{id}/edit", s.editSchedule)
	r.Post("/schedules/{id}/delete", s.deleteSchedule)
	r.Get("/schedules", s.schedules)
}

// GET request for /schedule: shows default schedule for current user
func (s Server) schedule(w http.ResponseWriter, r *http.Request) {
	_ = pages.Schedule([]models.Period{}).Render(r.Context(), w)
}

// GET request to /schedules/create
func (s Server) createScheduleForm(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	emptySchedule := models.Schedule{}
	validation := schedule.Validate(emptySchedule)

	teachers, _ := s.Teachers.List(context)
	periods, _ := s.Periods.List(context)
	_ = pages.CreateSchedule(emptySchedule, teachers, periods, validation, nil).Render(context, w)
}

// POST request to /schedules/create/validate
func (s Server) validateCreateSchedule(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
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

	teachers, _ := s.Teachers.List(context)
	for _, teacher := range teachers {
		if teacher.ID == signals.Schedule.TeacherID {
			selectedTeacher = &teacher
		}
	}
	periods, _ := s.Periods.List(context)
	validation := schedule.Validate(model)
	patchTempl(w, r, blocks.CreateScheduleForm(model, teachers, periods, validation, selectedTeacher))
}

// POST request to /schedules/create
func (s Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := schedule.CreateScheduleCommandHandler(r.Context(), schedule.CreateScheduleCommand{
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
func (s Server) getEditSchedule(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Schedule models.ScheduleSignals `json:"schedule"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(context, scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.Schedule{
		Id:        scheduleID,
		Title:     signals.Schedule.Title,
		TeacherId: signals.Schedule.TeacherID,
	}
	selectedTeacher := ""
	// if the user has selected a teacher, use that as the default selection
	// otherwise use the schedule's teacher ID data
	if signals.Schedule.TeacherID != "" && scheduleRes.TeacherId != signals.Schedule.TeacherID {
		selectedTeacher = signals.Schedule.TeacherID
	} else {
		selectedTeacher = scheduleRes.TeacherId
	}

	validation := schedule.Validate(model)
	teachers, err := s.Teachers.List(context)
	periods, err := s.Periods.List(context)
	_ = pages.EditSchedule(*scheduleRes, teachers, periods, validation, selectedTeacher).Render(context, w)
}

func (s Server) editScheduleStream(w http.ResponseWriter, r *http.Request) {
	//sse := newSSE(w, r)
	ctx := r.Context()
	scheduleID := chi.URLParam(r, "id")

	updates := make(chan struct{}, 1)
	notify := func() {
		select {
		case updates <- struct{}{}:
		default:
		}
	}

	// refreshes the view state for initial population of value
	if err := s.refreshScheduleViewState(ctx, scheduleID); err != nil {
		println(err.Error())
		return
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

	sub, err := s.Subscriber.Subscribe(ctx, schedule.Channel("idk"), func(context.Context, []byte) {
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
			var schedule models.Schedule
			if err := entry.JSON(&schedule); err != nil {
				println(err.Error())
				return
			}
			println(schedule.Title)
		}
	}
}

// POST request to /schedules/{id}/edit/validate
func (s Server) validateEditSchedule(w http.ResponseWriter, r *http.Request) {
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
	teachers, _ := s.Teachers.List(ctx)
	periods, _ := s.Periods.List(ctx)
	validation := schedule.Validate(model)
	viewstore.PutState(ctx, s.ViewStore, scheduleID, model)
	patchTempl(w, r, blocks.EditScheduleForm(model, teachers, periods, validation, signals.Schedule.TeacherID))
}

// POST request to /schedules/{id}/edit
func (s Server) editSchedule(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Title     string   `json:"title"`
		TeacherId string   `json:"teacher_id"`
		PeriodIDs []string `json:"period_ids"`
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scheduleID := chi.URLParam(r, "id")
	currentPeriodIDs, _ := s.Schedules.ListSchedulePeriodIDs(context, scheduleID)
	proposedPeriodIDs := signals.PeriodIDs
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
				println("remove: ", periodID)
				result, err := schedule.RemovePeriodFromScheduleCommandHandler(context, schedule.RemovePeriodFromScheduleCommand{
					ScheduleID: scheduleID,
					PeriodID: periodID,
					Metadata: eventstore.HTTPCommandMetadata(r),
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
		for _, periodID := range proposedPeriodIDs {
			if !currentMap[periodID] {
				println("add: ", periodID)
				result, err := schedule.PeriodAddedToScheduleCommandHandler(context, schedule.PeriodAddedToScheduleCommand{
					ScheduleID: scheduleID,
					PeriodID: periodID,
					Metadata: eventstore.HTTPCommandMetadata(r),
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
	// _, err = schedule.UpdateScheduleCommandHandler(context, schedule.UpdateScheduleCommand{
	// 	Id:        scheduleID,
	// 	Title:     signals.Title,
	// 	TeacherId: signals.TeacherId,
	// 	Metadata:  eventstore.HTTPCommandMetadata(r),
	// }, s.EventSaver, s.EventRetriever)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
}

// POST request to /schedules/{id}/delete
func (s Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "id")
	_, err := schedule.DeleteScheduleCommandHandler(r.Context(), schedule.DeleteScheduleCommand{
		ScheduleID: scheduleID,
		Metadata:   eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}

// GET request to /schedules
func (s Server) schedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := s.Schedules.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.Schedules(schedules).Render(r.Context(), w)
}

func (s Server) refreshScheduleViewState(ctx context.Context, scheduleID string) error {
 schedule, err := s.Schedules.Get(ctx, scheduleID)
 if err != nil {
	return err
 }
 return viewstore.PutState(ctx, s.ViewStore, scheduleID, schedule)
}