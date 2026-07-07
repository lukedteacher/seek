package httpserver

import (
	"net/http"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/views/blocks"
	"seek/internal/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get("/periods", s.getPeriods)
	r.Get("/periods/create", s.getPeriodCreate)
	r.Post("/periods/create/validate", s.postPeriodCreateValidate)
	r.Post("/periods/create", s.postPeriodCreate)
	r.Get("/periods/{id}", s.getPeriodView)
	r.Get("/periods/{id}/edit", s.getPeriodEdit)
	r.Post("/periods/{id}/edit", s.editPeriod)
	r.Post("/periods/{id}/edit/validate", s.postPeriodEditValidate)
	r.Delete("/periods/{id}", s.deletePeriod)
}

// GET request to /periods
// lists all periods
// probably won't need this later
func (s Server) getPeriods(w http.ResponseWriter, r *http.Request) {
	// defines the view (card or list or... something else?)
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

// GET request to /periods/create
// TODO figure out if there's a way to have this use the same form as edit?
func (s Server) getPeriodCreate(w http.ResponseWriter, r *http.Request) {
	validation := period.Validate(models.Period{})
	_ = pages.CreatePeriod(validation).Render(r.Context(), w)
}

// POST request to /periods/create/validate
func (s Server) postPeriodCreateValidate(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Period models.PeriodSignals `json:"period"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vcp signal read: ", err.Error())
		return
	}

	daysBitmask := models.DaysSignalsToDaysBitmask(signals.Period.Days)
	model := models.Period{
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      daysBitmask,
	}
	validation := period.Validate(model)
	patchTempl(w, r, blocks.CreatePeriodForm(validation))
}

// POST request to /periods/create
func (s Server) postPeriodCreate(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Period models.PeriodSignals `json:"period"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("pc signal read: ", err.Error())
		return
	}
	daysBitmask := models.DaysSignalsToDaysBitmask(signals.Period.Days)
	model := models.Period{
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      daysBitmask,
	}
	validation := period.Validate(model)
	if validation == nil {
		println("some error")
	}
	command := period.CreatePeriodCommand{
		Title:     model.Title,
		StartTime: model.StartTime,
		Duration:  model.Duration,
		Days:      daysBitmask,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}
	_, err := period.CreatePeriodCommandHandler(r.Context(), command, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}

	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&Signals{}, sse)
	})
}

// GET request to /periods/{id}
func (s Server) getPeriodView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	pm, err := s.Periods.Get(r.Context(), periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pm == nil {
		_ = pages.PeriodNotFound().Render(ctx, w)
	} else {
		_ = pages.PeriodView(*pm).Render(ctx, w)
	}
}

// GET request to /periods/{id}/edit
func (s Server) getPeriodEdit(w http.ResponseWriter, r *http.Request) {
	periodID := chi.URLParam(r, "id")
	pm, err := s.Periods.Get(r.Context(), periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	daysSignals := models.DaysBitmaskToDaysSignals(pm.Days)
	signals := models.PeriodSignals{
		ID:        periodID,
		Title:     pm.Title,
		StartTime: pm.StartTime,
		Duration:  int(pm.Duration),
		Days:      daysSignals,
	}
	
	validation := period.Validate(*pm)
	_ = pages.EditPeriod(signals, signals, validation).Render(r.Context(), w)
}

// POST request to /periods/{id}/edit/validate
func (s Server) postPeriodEditValidate(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Period models.PeriodSignals `json:"period"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("vep signals: ", err.Error())
		return
	}

	signals.Period.ID = chi.URLParam(r, "id")
	daysBitmask := models.DaysSignalsToDaysBitmask(signals.Period.Days)

	model := models.Period{
		Id:        signals.Period.ID,
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      daysBitmask,
	}
	validation := period.Validate(model)
	patchTempl(w, r, blocks.EditPeriodForm(signals.Period, signals.Period , validation))
}

// POST request to /periods/{id}/edit
func (s Server) editPeriod(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		Period models.PeriodSignals `json:"period"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println("ep signals: ", err.Error())
		return
	}

	daysBitmask := models.DaysSignalsToDaysBitmask(signals.Period.Days)

	updatePeriodCommand := period.UpdatePeriodCommand{
		Id:        chi.URLParam(r, "id"),
		Title:     signals.Period.Title,
		StartTime: signals.Period.StartTime,
		Duration:  int64(signals.Period.Duration),
		Days:      daysBitmask,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}

	result, err := period.UpdatePeriodCommandHandler(r.Context(), updatePeriodCommand, s.EventSaver, s.EventRetriever)
	if err != nil {
		println("upch e: ", err.Error())
		return
	}
	if result.Skipped == true {
		println("update skipped")
		return
	}
}

// DELETE request to /periods/{id}
func (s Server) deletePeriod(w http.ResponseWriter, r *http.Request) {
	periodID := chi.URLParam(r, "id")
	_, err := period.DeletePeriodCommandHandler(r.Context(), period.DeletePeriodCommand{
		PeriodID: periodID,
		Metadata: eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
