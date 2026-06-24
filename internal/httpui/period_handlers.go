package httpui

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get("/periods", s.periods)
	r.Get("/period/create", s.createPeriodForm)
	r.Post("/period/create", s.createPeriod)
	r.Post("/period/create/validate", s.validateCreatePeriod)
	r.Get("/period/{id}", s.editPeriodForm)
	r.Get("/period/{id}/edit", s.editPeriodForm)
	r.Post("/period/{id}/edit", s.editPeriod)
	r.Post("/period/{id}/edit/validate", s.validateEditPeriod)
}

func (s Server) periods(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		View int64 `json:"view"`
	}
	signals := &Signals{}
	periods, err := s.Periods.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.ReadSignals(r, signals)
	
	_ = pages.Periods(signals.View, periods).Render(r.Context(), w)
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

// GET request to /period/{id}
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
	periodRes, err := s.Periods.Get(r.Context(), periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.Period{
		Title: signals.Title,
		StartTime: signals.StartTime,
		Duration: signals.Duration,
		Days: signals.Duration,
	}
	validation := period.Validate(model)
	_ = pages.EditPeriod(*periodRes, validation).Render(r.Context(), w)
}

// POST request to /period/{id}/edit
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

// POST request to /period/create/validate
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

// POST request to /period/{id}/edit/validate
func (s Server) validateEditPeriod(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "id")
	model := models.Period{
		Id: id,
		Title: signals.Title,
		StartTime: signals.StartTime,
		Duration: signals.Duration,
		Days: signals.Duration,
	}
	validation := period.Validate(model)
	patchTempl(w, r, pages.EditPeriod(model, validation))
}