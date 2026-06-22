package httpui

import (
	"net/http"

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
	_ = pages.CreatePeriod().Render(r.Context(), w)
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
    Title:    signals.Title,
    StartTime:     signals.StartTime,
    Duration:        signals.Duration,
    Days:     signals.Days,
    Metadata:     eventstore.HTTPCommandMetadata(r),
  }, s.EventSaver)
  if err != nil {
		println("error")
    writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
      return flashError(sse, err.Error())
    })
    return
  }
}