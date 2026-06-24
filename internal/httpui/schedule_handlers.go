package httpui

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/schedule"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedule", s.schedule)
	r.Get("/schedule/create", s.createScheduleForm)
	r.Post("/schedule/create", s.createSchedule)
	r.Post("/schedule/create/validate", s.validateCreateSchedule)
	r.Get("/schedule/{id}", s.editScheduleForm)
	r.Get("/schedule/{id}/edit", s.editScheduleForm)
	r.Post("/schedule/{id}/edit", s.editSchedule)
	r.Post("/schedule/{id}/edit/validate", s.validateEditSchedule)
}

// GET request for /schedule: shows default schedule for current user
func (s Server) schedule(w http.ResponseWriter, r *http.Request) {
	_ = pages.Schedule().Render(r.Context(), w)
}

func (s Server) scheduleStream(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	ctx := r.Context()

	watcher, err := s.ViewStore.Watch(ctx, "schedule", viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		_ = alert(sse, err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watcher.Updates():
			if !ok {
				return
			}
			page := pages.Schedule()
			if err := sse.PatchElementTempl(page); err != nil {
				return
			}
		}
	}
}

// get request to /schedule/create
func (s Server) createScheduleForm(w http.ResponseWriter, r *http.Request) {
	validation := schedule.Validate(models.Schedule{})
	_ = pages.CreateSchedule(validation).Render(r.Context(), w)
}

// post request to /schedule
func (s Server) createSchedule(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title    string `json:"title"`
    TeacherId   string `json:"teacher_id"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }

  _, err := schedule.CreateScheduleCommandHandler(r.Context(), schedule.CreateScheduleCommand{
    Title:     signals.Title,
    TeacherId: signals.TeacherId,
    Metadata:  eventstore.HTTPCommandMetadata(r),
  }, s.EventSaver)
  if err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
}

// GET request to /schedule/{id}
func (s Server) editScheduleForm(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    TeacherId string `json:"teacher_id"`
    Duration  int64  `json:"duration"`
    Days      int64	 `json:"days"`
  }
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }

	scheduleID := chi.URLParam(r, "id")
	scheduleRes, err := s.Schedules.Get(r.Context(), scheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.Schedule{
		Title: signals.Title,
		TeacherId: signals.TeacherId,
	}
	validation := schedule.Validate(model)
	_ = pages.EditSchedule(*scheduleRes, validation).Render(r.Context(), w)
}

// POST request to /schedule/{id}/edit
func (s Server) editSchedule(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    TeacherId string `json:"teacher_id"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
	
	scheduleID := chi.URLParam(r, "id")
  
  result, err := schedule.UpdateScheduleCommandHandler(r.Context(), schedule.UpdateScheduleCommand{
		Id:        scheduleID,
    Title:     signals.Title,
    TeacherId: signals.TeacherId,
    Metadata:  eventstore.HTTPCommandMetadata(r),
  }, s.EventSaver, s.EventRetriever)
  if err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
	if result.Skipped == true {
		println("update skipped")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
  
  writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
    return clearNewStudentForm(sse)
  })
}

// POST request to /schedule/create/validate
func (s Server) validateCreateSchedule(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    TeacherId string `json:"teacher_id"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
	model := models.Schedule{
		Title: signals.Title,
		TeacherId: signals.TeacherId,
	}
	validation := schedule.Validate(model)
	patchTempl(w, r, pages.CreateSchedule(validation))
}

// POST request to /schedule/{id}/edit/validate
func (s Server) validateEditSchedule(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    TeacherId string `json:"teacher_id"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
	id := chi.URLParam(r, "id")
	model := models.Schedule{
		Id: id,
		Title: signals.Title,
		TeacherId: signals.TeacherId,
	}
	validation := schedule.Validate(model)
	patchTempl(w, r, pages.EditSchedule(model, validation))
}