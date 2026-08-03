package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"seek/internal/eventstore"
	epevents "seek/internal/features/educators_periods/events"
	"seek/internal/features/periods/dto"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	"seek/internal/features/periods/pages"
	sdto "seek/internal/features/students/dto"
	spevents "seek/internal/features/students_periods/events"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) periodRoutes(r chi.Router) {
	r.Get("/periods", s.getPeriodsList)
	r.Get("/periods/stream", s.getPeriodsListStream)
	r.Get("/periods/create", s.getPeriodCreate)
	r.Get("/periods/create/stream", s.getPeriodCreateStream)
	r.Post("/periods/create/validate", s.postPeriodCreateValidate)
	r.Post("/periods/create/validate/{field}", s.postPeriodCreateValidateField)
	r.Post("/periods/create", s.postPeriodCreate)
	r.Get("/periods/{id}", s.getPeriodView)
	r.Get("/periods/{id}/stream", s.getPeriodViewStream)
	r.Get("/periods/{id}/edit", s.getPeriodEdit)
	r.Get("/periods/{id}/edit/stream", s.getPeriodEditStream)
	r.Post("/periods/{id}/edit/validate", s.postPeriodEditValidate)
	r.Post("/periods/{id}/edit", s.postPeriodEdit)
	r.Post("/periods/{id}/archive", s.postPeriodArchive)
	r.Delete("/periods/{id}", s.deletePeriod)
}

// GET request to /periods
func (s Server) getPeriodsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periods, err := s.ReadModels.Periods.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := dto.NewPeriodTableView(periods)
	view.URL = "/periods"
	_ = pages.List(user, view).Render(ctx, w)
}

// GET request to /periods/stream
func (s Server) getPeriodsListStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to any periods
	sub, err := s.Subscriber.Subscribe(ctx, events.ChannelAll(), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "periods list stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			// for now just reloads the page
			// consider adding a view store for the list
			periods, err := s.ReadModels.Periods.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "periods list stream db list", "err", err)
				return
			}
			view := dto.NewPeriodTableView(periods)
			view.URL = "/periods"
			sse.PatchElementTempl(pages.List(user, view))
		}
	}
}

// GET request to /periods/create
func (s Server) getPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)

	// create an empty model
	empty, _ := models.NewPeriod()

	// list all educators
	educators, err := s.ReadModels.Educators.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get period create db list educators", "err", err)
		return
	}

	// list all students
	students, err := s.ReadModels.Students.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get period create db list students", "err", err)
		return
	}

	// create the form view
	view := dto.NewPeriodFormView(empty, students, educators)

	// create blank views for the schedule
	psvs := []dto.PeriodScheduleView{}

	// set the URL
	view.URL = "/periods/create"
	_ = pages.Create(user, view, psvs).Render(ctx, w)
}

// GET request to /periods/create/stream
func (s Server) getPeriodCreateStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)

	// watches the kv "seek-view-state" for create view state changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		"new",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "period create stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			signals := &struct {
				Period dto.PeriodFormView `json:"period"`
			}{}
			if err := entry.JSON(signals); err != nil {
				s.Logger.Error("get period create json", "err", err)
				return
			}

			// convert the form view to a model
			model := signals.Period.ToPeriod()

			// list all educators
			educators, err := s.ReadModels.Educators.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "get period create db list educators", "err", err)
				return
			}

			// get student list
			students, err := s.ReadModels.Students.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "get period create stream db list students by iep service", "err", err)
				return
			}

			// create the form view
			view := dto.NewPeriodFormView(&model, students, educators)

			// create views for the schedule
			psvs := dto.NewPeriodScheduleViews(model)

			// set the URL in the view
			view.URL = "/periods/create"
			sse.PatchElementTempl(pages.Create(user, view, psvs))
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
		s.Logger.ErrorContext(ctx, "period create validate signals", "err", err.Error())
		return
	}

	// saves the view in a nats kv so the SSE can update
	if err := viewstore.PutState(ctx, s.ViewStore, "new", signals); err != nil {
		s.Logger.ErrorContext(ctx, "post period create validate viewstore", "err", err)
	}
}

// POST request to /periods/create/validate/{field}
func (s Server) postPeriodCreateValidateField(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "pcvf signal read", "error", err)
		return
	}
	field := chi.URLParam(r, "field")
	s.Logger.DebugContext(ctx, "validating field", "field", field, "st", signals.Period.ServiceType)

	// convert the form view to a temporary Period for business logic
	temp := signals.Period.ToPeriod()

	switch field {
	case "starttime":
		// update start time → recalculate end time using current duration
		temp.UpdateStartTime(signals.Period.StartTime)
		// copy back the recalculated end time (duration unchanged)
		signals.Period.EndTime = temp.EndTime
		s.Logger.DebugContext(ctx, "start time updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

	case "endtime":
		// update end time → recalculate duration using current start time
		temp.UpdateEndTime(signals.Period.EndTime)
		// copy back the new duration (end time already set)
		signals.Period.Duration = temp.Duration
		signals.Period.StartTime = temp.StartTime
		s.Logger.DebugContext(ctx, "end time updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

	case "duration":
		// update duration → recalculate end time using current start time
		temp.UpdateDuration(signals.Period.Duration)
		// copy back the recalculated end time
		signals.Period.EndTime = temp.EndTime
		s.Logger.DebugContext(ctx, "duration updated", "start", temp.StartTime, "end", temp.EndTime, "duration", temp.Duration)

	default:
		s.Logger.WarnContext(ctx, "unknown validation field", "field", field)
		http.Error(w, "unknown field", http.StatusBadRequest)
		return
	}

	// save the updated signals to the view store so the SSE can refresh the form
	if err := viewstore.PutState(ctx, s.ViewStore, "new", signals); err != nil {
		s.Logger.ErrorContext(ctx, "view store error", "error", err)
	}
}

// POST request to /periods/create
func (s Server) postPeriodCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post period create signals", "err", err)
		return
	}
	command := events.CreatePeriodCommand{
		Title:       signals.Period.Title,
		ServiceType: signals.Period.ServiceType,
		StartTime:   signals.Period.StartTime,
		Duration:    signals.Period.Duration,
		DaysBitmask: signals.Period.Days.ToBitmask(),
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.CreatePeriodCommandHandler(ctx, command, s.EventSaver)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post period create command handler", "err", err)
		return
	}

	// create period student associations
	studentIDs := strings.Split(signals.Period.StudentIDs, ",")
	for _, studentID := range studentIDs {
		periodStudentAddCommand := spevents.AddStudentToPeriodCommand{
			PeriodID:  result.EventID,
			StudentID: studentID,
			Metadata:  eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		_, err := spevents.AddStudentToPeriodCommandHandler(
			ctx,
			periodStudentAddCommand,
			s.EventSaver,
			s.EventRetriever,
		)
		if err != nil {
			s.Logger.ErrorContext(ctx, "post period create command handler", "err", err)
		}
	}
	// redirect to the view page after creating
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/periods/%s", result.EventID))
}

// GET request to /periods/{id}
func (s Server) getPeriodView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	model, err := s.ReadModels.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if model == nil {
		_ = pages.NotFound(user).Render(ctx, w)
		return
	}
	view := dto.NewPeriodView(model)
	view.URL = fmt.Sprintf("/periods/%s", periodID)
	studentIDs, _ := s.ReadModels.StudentPeriods.ListStudentIDsForPeriod(ctx, model.ID)
	for i := range studentIDs {
		student, _ := s.ReadModels.Students.GetByID(ctx, studentIDs[i])
		studentView := sdto.NewStudentView(student)
		view.Students = append(view.Students, studentView)
	}
	_ = pages.View(user, view).Render(ctx, w)
}

// GET request to /periods/{id}/stream
func (s Server) getPeriodViewStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to the underlying model
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "get period view stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	if err := s.refreshPeriodViewState(ctx, periodID); err != nil {
		s.Logger.ErrorContext(ctx, "get period view stream refresh", "err", err)
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
		s.Logger.ErrorContext(ctx, "get period view stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal(): // triggers when the read model publishes
			if err := s.refreshPeriodViewState(ctx, periodID); err != nil {
				if err.Error() == "period not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				s.Logger.ErrorContext(ctx, "get period view stream refresh in select", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			model := &models.Period{}
			if err := entry.JSON(model); err != nil {
				s.Logger.ErrorContext(ctx, "get period view stream json", "err", err)
				return
			}
			view := dto.NewPeriodView(model)
			view.URL = fmt.Sprintf("/periods/%s", periodID)
			studentIDs, _ := s.ReadModels.StudentPeriods.ListStudentIDsForPeriod(ctx, model.ID)
			for i := range studentIDs {
				student, _ := s.ReadModels.Students.GetByID(ctx, studentIDs[i])
				studentView := sdto.NewStudentView(student)
				view.Students = append(view.Students, studentView)
			}
			sse.PatchElementTempl(pages.View(user, view))
		}
	}
}

// GET request to /periods/{id}/edit
func (s Server) getPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")

	// get period model
	period, err := s.ReadModels.Periods.Get(ctx, periodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set educatorIDs for model (because they are not loaded in the get function... TODO change that?)
	selectedEducators, _ := s.ReadModels.EducatorPeriods.ListEducatorIDsForPeriod(ctx, period.ID)
	period.EducatorIDs = selectedEducators

	// set studentIDs for model (because they are not loaded in the get function... TODO change that?)
	selectedStudents, _ := s.ReadModels.StudentPeriods.ListStudentIDsForPeriod(ctx, period.ID)
	period.StudentIDs = selectedStudents

	// get educators and make views
	educators, err := s.ReadModels.Educators.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "period edit db list educators", "err", err)
		return
	}

	// get students
	students, err := s.ReadModels.Students.List(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "period edit db list students", "err", err)
		return
	}

	// create the form view
	view := dto.NewPeriodFormView(period, students, educators)

	// set the URL in the view
	view.URL = fmt.Sprintf("/periods/%s/edit", periodID)
	_ = pages.Edit(user, view, []dto.PeriodScheduleView{}).Render(ctx, w)
}

// GET request to /period/{id}/stream
func (s Server) getPeriodEditStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	sse := newSSE(w, r)

	// subscribes to the channel which publishes changes to the underlying model
	notifier := NewDedupeNotifier()
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(periodID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "period edit stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	// watches the kv "seek-view-state" for edit view state changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		periodID+".edit",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "edit view stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			// triggers when there is a change published on the read model's channel
			if err := s.refreshPeriodEditState(ctx, periodID); err != nil {
				// in case the period no longer exists
				if err.Error() == "period not found" {
					sse.PatchElementTempl(pages.NotFound(user))
				}
				s.Logger.ErrorContext(ctx, "period edit stream refresh in select", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			periodFormView := &struct {
				Period dto.PeriodFormView `json:"period"`
			}{}
			if err := entry.JSON(periodFormView); err != nil {
				s.Logger.Error("period edit stream json", "err", err)
				return
			}

			// convert the form view to a model
			period := periodFormView.Period.ToPeriod()

			// get the periods for the selected students
			periods := make([]models.Period, 0)
			periods = append(periods, period)
			for _, studentID := range period.StudentIDs {
				s.Logger.Debug("period edit stream", "student ID", studentID)
				studentPeriods, _ := s.ReadModels.Periods.ListPeriodsForStudent(ctx, studentID)
				periods = append(periods, studentPeriods...)
			}

			// get educators
			educators, err := s.ReadModels.Educators.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "period edit stream db list educators", "err", err)
				return
			}

			// get students
			students, err := s.ReadModels.Students.List(ctx)
			if err != nil {
				s.Logger.ErrorContext(ctx, "period edit stream db list students", "err", err)
				return
			}

			// create the form view
			newPeriodFormView := dto.NewPeriodFormView(&period, students, educators)

			// create views for the schedule
			psvs := dto.NewPeriodScheduleViews(periods...)
			s.Logger.Debug("edit SSE", "sv length", len(psvs))
			// set the URL in the view
			newPeriodFormView.URL = fmt.Sprintf("/periods/%s/edit", periodID)
			sse.PatchElementTempl(pages.Edit(user, newPeriodFormView, psvs))
		}
	}
}

// POST request to /periods/{id}/edit/validate
func (s Server) postPeriodEditValidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "id")
	signals := &struct {
		FormView dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "period edit validate signals", "err", err)
		return
	}
	// ensures signal gets the ID
	signals.FormView.ID = periodID

	// saves the view in a nats kv so the SSE can update
	if err := viewstore.PutState(ctx, s.ViewStore, periodID+".edit", signals); err != nil {
		s.Logger.ErrorContext(ctx, "post period create validate viewstore", "err", err)
	}
}

// POST request to /periods/{id}/edit
func (s Server) postPeriodEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	signals := &struct {
		Period dto.PeriodFormView `json:"period"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		s.Logger.ErrorContext(ctx, "post period edit signals", "err", err)
		return
	}
	periodID := chi.URLParam(r, "id")
	command := events.UpdatePeriodCommand{
		ID:          periodID,
		Title:       signals.Period.Title,
		ServiceType: signals.Period.ServiceType,
		StartTime:   signals.Period.StartTime,
		Duration:    signals.Period.Duration,
		DaysBitmask: signals.Period.Days.ToBitmask(),
		Metadata:    eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}
	result, err := events.UpdatePeriodCommandHandler(ctx, command, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post period edit command handler", "err", err)
		return
	}
	if result.Skipped == true {
		s.Logger.InfoContext(ctx, "post period edit command handler", "skipped", result.Skipped)
	}

	// sync the educators
	secmd := epevents.SyncEducatorsInPeriodCommand{
		PeriodID:            periodID,
		ProposedEducatorIDs: strings.Split(signals.Period.EducatorIDs, ","),
	}
	_, err = epevents.SyncEducatorsInPeriodCommandHandler(ctx, secmd, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post period edit sync educators", "err", err)
	}

	// sync the students
	spcmd := spevents.SyncStudentsInPeriodCommand{
		PeriodID:           periodID,
		ProposedStudentIDs: strings.Split(signals.Period.StudentIDs, ","),
	}
	_, err = spevents.SyncStudentsInPeriodCommandHandler(ctx, spcmd, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "post period edit sync students", "err", err)
	}

	// redirect the view page after editing
	sse := newSSE(w, r)
	sse.Redirect(fmt.Sprintf("/periods/%s", periodID))
}

// POST request to /periods/{id}/archive
func (s Server) postPeriodArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	_, err := events.ArchivePeriodCommandHandler(ctx, events.ArchivePeriodCommand{
		PeriodID: periodID,
		Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "archive period command handler", "err", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/periods")
}

// DELETE request to /periods/{id}
func (s Server) deletePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	periodID := chi.URLParam(r, "id")
	_, err := events.DeletePeriodCommandHandler(ctx, events.DeletePeriodCommand{
		PeriodID: periodID,
		Metadata: eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		s.Logger.ErrorContext(ctx, "delete period command handler", "err", err)
		return
	}
	sse := newSSE(w, r)
	sse.Redirect("/periods")
}

func (s Server) refreshPeriodViewState(ctx context.Context, periodID string) error {
	period, err := s.ReadModels.Periods.Get(ctx, periodID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, period.ID+".view", period)
}

func (s Server) refreshPeriodEditState(ctx context.Context, periodID string) error {
	period, err := s.ReadModels.Periods.Get(ctx, periodID)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, period.ID+".edit", period)
}
