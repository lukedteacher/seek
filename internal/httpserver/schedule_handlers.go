package httpserver

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/schedule"
	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedule", s.schedule)
	r.Get("/schedules/create", s.createScheduleForm)
	r.Post("/schedules/create/validate", s.validateCreateSchedule)
	r.Post("/schedules/create", s.createSchedule)
	r.Get("/schedules/{id}", s.editScheduleForm)
	r.Get("/schedules/{id}/edit", s.editScheduleForm)
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
	_ = pages.CreateSchedule(emptySchedule, teachers, periods, validation, "select a teacher").Render(context, w)
}

// POST request to /schedules/create/validate
func (s Server) validateCreateSchedule(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Title     string   `json:"title"`
		TeacherId string   `json:"teacher_id"`
		Periods   []string `json:"periods"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := models.Schedule{
		Title:     signals.Title,
		TeacherId: signals.TeacherId,
	}
	selectedTeacher := signals.TeacherId

	teachers, _ := s.Teachers.List(context)
	periods, _ := s.Periods.List(context)
	validation := schedule.Validate(model)
	_ = pages.CreateSchedule(model, teachers, periods, validation, selectedTeacher).Render(context, w)
}

// POST request to /schedules/create
func (s Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Title     string `json:"title"`
		TeacherId string `json:"teacher_id"`
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := schedule.CreateScheduleCommandHandler(r.Context(), schedule.CreateScheduleCommand{
		Title:     signals.Title,
		TeacherId: signals.TeacherId,
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
func (s Server) editScheduleForm(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Title     string   `json:"title"`
		TeacherId string   `json:"teacher_id"`
		Periods   []string `json:"periods"`
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
		Title:     signals.Title,
		TeacherId: signals.TeacherId,
	}
	selectedTeacher := ""
	// if the user has selected a teacher, use that as the default selection
	// otherwise use the schedule's teacher ID data
	if signals.TeacherId != "" && scheduleRes.TeacherId != signals.TeacherId {
		selectedTeacher = signals.TeacherId
	} else {
		selectedTeacher = scheduleRes.TeacherId
	}

	validation := schedule.Validate(model)
	teachers, err := s.Teachers.List(context)
	periods, err := s.Periods.List(context)
	_ = pages.EditSchedule(*scheduleRes, teachers, periods, validation, selectedTeacher).Render(context, w)
}

// POST request to /schedules/{id}/edit/validate
func (s Server) validateEditSchedule(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		Title     string   `json:"title"`
		TeacherId string   `json:"teacher_id"`
		Periods   []string `json:"periods"`
	}

	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := chi.URLParam(r, "id")
	model := models.Schedule{
		Id:        id,
		Title:     signals.Title,
		TeacherId: signals.TeacherId,
	}
	teachers, _ := s.Teachers.List(context)
	periods, _ := s.Periods.List(context)
	validation := schedule.Validate(model)
	patchTempl(w, r, pages.EditSchedule(model, teachers, periods, validation, signals.TeacherId))
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

	// if len(signals.PeriodIDs) > 0 {
	// 	for _, period := range signals.PeriodIDs {
	// 		_, err = schedule.PeriodAddedToScheduleCommandHandler(context, schedule.PeriodAddedToScheduleCommand{
	// 			ScheduleID: scheduleID,
	// 			PeriodID:   period,
	// 			Metadata:   eventstore.HTTPCommandMetadata(r),
	// 		}, s.EventSaver, s.EventRetriever)
	// 		if err != nil {
	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
	// 			return
	// 		}
	// 	}
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
