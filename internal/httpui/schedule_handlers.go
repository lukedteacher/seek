package httpui

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) scheduleRoutes(r chi.Router) {
	r.Get("/schedule", s.schedule)
	r.Get("/periods/create", s.createPeriodForm)
	r.Post("/periods/create", s.createPeriod)
	r.Post("/periods/create/validate", s.validateCreatePeriod)
	r.Get("/periods/{id}", s.editPeriodForm)
	r.Patch("/periods/{id}", s.editPeriod)
}

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

// get request to /period/create
func (s Server) createPeriodForm(w http.ResponseWriter, r *http.Request) {
	validation := period.Validate(models.Period{})
	_ = pages.CreatePeriod(validation).Render(r.Context(), w)
}

// post request to /period
func (s Server) createPeriod(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title    string `json:"title"`
    StartTime   string `json:"start_time"`
    Duration    int64 `json:"duration"`
    Days        int64	`json:"days"`
  }
  
  signals := &Signals{}
  if err := datastar.ReadSignals(r, signals); err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }

  _, err := period.CreatePeriodCommandHandler(r.Context(), period.CreatePeriodCommand{
    Title:     signals.Title,
    StartTime: signals.StartTime,
    Duration:  signals.Duration,
    Days:      signals.Days,
    Metadata:  eventstore.HTTPCommandMetadata(r),
  }, s.EventSaver)
  if err != nil {
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
}

// get request to /period/{id}
func (s Server) editPeriodForm(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    StartTime string `json:"start_time"`
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

	periodID := chi.URLParam(r, "id")
	period, err := s.Periods.Get(r.Context(), periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.EditPeriod(*period).Render(r.Context(), w)
}

// patch request to /periods/{id}
func (s Server) editPeriod(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    StartTime string `json:"start_time"`
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
	periodID := chi.URLParam(r, "id")
  
  result, err := period.UpdatePeriodCommandHandler(r.Context(), period.UpdatePeriodCommand{
		Id:        periodID,
    Title:     signals.Title,
    StartTime: signals.StartTime,
    Duration:  signals.Duration,
    Days:      signals.Days,
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

// patch request to /periods/{id}
func (s Server) validateCreatePeriod(w http.ResponseWriter, r *http.Request) {
  type Signals struct {
    Title     string `json:"title"`
    StartTime string `json:"start_time"`
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
	model := models.Period{
		Title: signals.Title,
		StartTime: signals.StartTime,
		Duration: signals.Duration,
		Days: signals.Duration,
	}
	validation := period.Validate(model)
	patchTempl(w, r, pages.CreatePeriod(validation))
}