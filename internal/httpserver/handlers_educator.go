package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"time"

	"seek/internal/errs"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/events"
	"seek/internal/features/educators/models"
	"seek/internal/features/educators/pages"
	periodEvents "seek/internal/features/periods/events"
	scheduleDTO "seek/internal/features/schedules/dto"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	studentModels "seek/internal/features/students/models"
	userModels "seek/internal/features/users/models"
	uifilters "seek/internal/ui/core/coreblocks/filters"
	"seek/internal/viewstore"
	"seek/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

type educatorSession struct {
	User     userModels.User
	Educator *models.Educator
	Key      string
	Notifier *DedupeNotifier
	Updates  <-chan viewstore.Entry
	sub      eventstore.MessageSubscription
	watcher  viewstore.Watcher
}

func (s *educatorSession) Close() {
	s.sub.Close()
	s.watcher.Stop()
}

func openEducatorSession(
	ctx context.Context,
	r *http.Request,
	username string,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	rm *events.ReadModel,
	kind string,
) (*educatorSession, error) {
	user := currentUser(r)
	if username == "" {
		return nil, fmt.Errorf("missing username: %w", errs.ErrNotFound)
	}

	view, err := rm.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get educator %q: %w", username, err)
	}

	notifier := NewDedupeNotifier()
	sub, err := subscriber.Subscribe(ctx, events.Channel(view.ID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe %q: %w", view.ID, err)
	}

	key, err := viewstore.Key(user.ID, kind, "educators", view.ID)
	if err != nil {
		sub.Close()
		return nil, fmt.Errorf("view key %q: %w", view.ID, err)
	}

	watcher, err := vs.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		sub.Close()
		return nil, fmt.Errorf("watch %q: %w", view.ID, err)
	}

	return &educatorSession{
		User:     user,
		Educator: &view.Educator,
		Key:      key,
		Notifier: notifier,
		Updates:  watcher.Updates(),
		sub:      sub,
		watcher:  watcher,
	}, nil
}

func refreshEducatorState[T any](
	ctx context.Context,
	vs viewstore.Store,
	session *educatorSession,
	username string,
	rm *events.ReadModel,
	build func(*models.Educator) T,
) error {
	view, err := rm.GetByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get educator %q: %w", username, err)
	}
	return viewstore.PutState(ctx, vs, session.Key, build(&view.Educator))
}

func (s Server) educatorRoutes(r chi.Router) {
	// educators list
	r.Get("/educators", getEducatorsList(s.Logger))
	r.Get("/educators/stream", getEducatorsListStream(s.Logger, s.ViewStore, s.ReadModels.Educators))
	r.Query("/educators", queryEducatorsList(s.Logger, s.ViewStore))
	// educator viewing
	r.Get("/educators/{username}", getEducatorView(s.Logger))
	r.Get("/educators/{username}/info", getEducatorViewInfo(s.Logger))
	r.Get("/educators/{username}/info/stream", getEducatorViewInfoStream(s.Logger, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Periods, s.ReadModels.Students))
	r.Get("/educators/{username}/schedule", getEducatorViewSchedule(s.Logger))
	r.Get("/educators/{username}/schedule/stream", getEducatorViewScheduleStream(s.Logger, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Periods, s.ReadModels.Students))
	r.Get("/educators/{username}/caseload", getEducatorViewCaseload(s.Logger))
	r.Get("/educators/{username}/caseload/stream", getEducatorViewCaseloadStream(s.Logger, s.ViewStore, s.ReadModels.Educators))
	// educator creation
	r.Get("/educators/create", getEducatorCreate(s.Logger))
	r.Get("/educators/create/stream", getEducatorCreateStream(s.Logger, s.ViewStore))
	r.Query("/educators/create/validate", queryEducatorFormValidate(s.Logger, s.ViewStore))
	r.Post("/educators/create", postEducatorCreate(s.Logger, s.EventSaver, s.ViewStore))
	// educator editing
	r.Get("/educators/{username}/edit", getEducatorEdit(s.Logger))
	r.Get("/educators/{username}/edit/stream", getEducatorEditStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators))
	r.Query("/educators/{username}/edit/validate", queryEducatorFormValidate(s.Logger, s.ViewStore))
	r.Post("/educators/{username}/edit", postEducatorEdit(s.Logger, s.EventSaver, s.EventRetriever, s.ViewStore))
	// other educator stuff
	r.Delete("/educators/{id}", deleteEducator(s.Logger, s.EventSaver, s.EventRetriever, s.ViewStore))
}

func getEducatorsList(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := dto.NewEducatorTableView([]models.Educator{})
		props := pages.ListPageProps{
			Signals: pages.ListPageSignals{},
			View:    view,
		}
		_ = pages.List(props).Render(ctx, w)
	}
}

func getEducatorsListStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return sseView(l, "educators list stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		user := currentUser(r)

		key, err := viewstore.Key(user.ID, "list", "educators", "")
		if err != nil {
			return fmt.Errorf("view key: %w", err)
		}

		pageURL := sharedmodels.GetURLFromContext(ctx)
		scrollAction := utils.Query(pageURL)

		const rowHeightPx = 44
		const bufferRows = 15

		render := func() (pages.ListPageProps, error) {
			signals, ok, err := viewstore.GetState[pages.ListPageSignals](ctx, vs, key)
			if err != nil {
				return pages.ListPageProps{}, fmt.Errorf("get state: %w", err)
			}
			if !ok {
				signals = pages.NewListPageSignals()
				if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
					return pages.ListPageProps{}, fmt.Errorf("seed state: %w", err)
				}
			}

			stringFilter := []string{}
			if len(signals.Filters.Roles) > 0 {
				stringFilter = models.ModelToString(signals.Filters.Roles)
			}
			sortCol := signals.Table.Sort.Column
			if sortCol == "first_l" {
				sortCol = "given_name"
			}

			listOpts := []events.ListOption{
				events.WithSortColumn(sortCol),
				events.WithSortDirection(signals.Table.Sort.Direction),
				events.WithGroupedBy(signals.Table.GroupedBy),
				events.WithSearchFilter(signals.Filters.Search),
				events.WithRoleFilter(stringFilter),
			}

			windowed := signals.ViewportHeight > 0
			startRow := 0
			if windowed {
				startRow = signals.ScrollTop/rowHeightPx - bufferRows
				if startRow < 0 {
					startRow = 0
				}
				visibleRows := signals.ViewportHeight/rowHeightPx + 2*bufferRows
				listOpts = append(listOpts, events.WithWindow(startRow, visibleRows))
			}

			result, err := educatorReadModel.ListWindowed(ctx, listOpts...)
			if err != nil {
				return pages.ListPageProps{}, fmt.Errorf("list educators: %w", err)
			}

			topSpacer := 0
			bottomSpacer := 0
			if windowed {
				topSpacer = startRow * rowHeightPx
				bottomSpacer = (result.Total - startRow - len(result.Educators)) * rowHeightPx
				if bottomSpacer < 0 {
					bottomSpacer = 0
				}
			}

			toastList, err := listToasts(ctx, vs, user.ID)
			if err != nil {
				return pages.ListPageProps{}, fmt.Errorf("list toasts: %w", err)
			}

			return pages.ListPageProps{
				Toasts:         toastList,
				Signals:        signals,
				View:           dto.NewEducatorTableViewWithSort(result.Educators, signals.Table.Columns, signals.Table.Sort, signals.Table.GroupedBy),
				ScrollAction:   scrollAction,
				TopSpacerPx:    topSpacer,
				BottomSpacerPx: bottomSpacer,
			}, nil
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastProps *pages.ListPageProps

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := pruneToasts(ctx, vs, user.ID); err != nil {
					l.ErrorContext(ctx, "tick: prune toasts", "err", err)
				}
				props, err := render()
				if err != nil {
					l.ErrorContext(ctx, "tick: render", "err", err)
					continue
				}
				if lastProps != nil && reflect.DeepEqual(*lastProps, props) {
					continue
				}
				if err := sse.PatchElementTempl(pages.List(props)); err != nil {
					l.ErrorContext(ctx, "tick: patch", "err", err)
					continue
				}
				lastProps = &props
			}
		}
	})
}

func queryEducatorsList(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.ListPageSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qel signals", "err", err)
			return
		}
		key, err := viewstore.Key(user.ID, "list", "educators", "")
		if err != nil {
			l.ErrorContext(ctx, "qel key", "err", err, "uid", user.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "qel vs put state", "err", err)
		}
	}
}

func getEducatorView(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		http.Redirect(w, r, fmt.Sprintf("/educators/%s/info", username), http.StatusFound)
	}
}

func getEducatorViewInfo(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Info(pages.InfoProps{}).Render(ctx, w)
	}
}

func getEducatorViewInfoStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
	periodReadModel *periodEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return sseView(l, "educator view info stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		username := chi.URLParam(r, "username")
		user := currentUser(r)

		render := func() (pages.InfoProps, error) {
			educator, err := educatorReadModel.GetByUsername(ctx, username)
			if err != nil {
				return pages.InfoProps{}, fmt.Errorf("get educator: %w", err)
			}

			periods, err := periodReadModel.ListPeriodsForEducator(ctx, educator.ID)
			if err != nil {
				return pages.InfoProps{}, fmt.Errorf("list periods: %w", err)
			}

			periodStudents := make(map[string][]studentModels.Student, len(periods))
			for _, period := range periods {
				periodModel, err := periodReadModel.GetWithIDs(ctx, period.ID)
				if err != nil {
					return pages.InfoProps{}, fmt.Errorf("get period %q: %w", period.ID, err)
				}
				if len(periodModel.StudentIDs) == 0 {
					continue
				}
				students, err := studentReadModel.ListByIDs(ctx, periodModel.StudentIDs)
				if err != nil {
					return pages.InfoProps{}, fmt.Errorf("get students for period %q: %w", period.ID, err)
				}
				periodStudents[period.ID] = students
			}

			var caseload []studentDTO.StudentView
			if educator.IsCaseManager() {
				caseManager, err := educatorReadModel.GetByUsernameWithCaseload(ctx, educator.Username)
				if err != nil {
					return pages.InfoProps{}, fmt.Errorf("get caseload: %w", err)
				}
				caseload = studentDTO.NewViews(caseManager.Caseload)
			}

			toastList, err := listToasts(ctx, vs, user.ID)
			if err != nil {
				return pages.InfoProps{}, fmt.Errorf("list toasts: %w", err)
			}

			return pages.InfoProps{
				Toasts:         toastList,
				Educator:       educator.Educator,
				Periods:        periods,
				PeriodStudents: periodStudents,
				Caseload:       caseload,
			}, nil
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastProps *pages.InfoProps

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := pruneToasts(ctx, vs, user.ID); err != nil {
					l.ErrorContext(ctx, "tick: prune toasts", "err", err)
				}
				props, err := render()
				if err != nil {
					if errors.Is(err, errs.ErrNotFound) {
						return err
					}
					l.ErrorContext(ctx, "tick: render", "err", err)
					continue
				}
				if lastProps != nil && reflect.DeepEqual(*lastProps, props) {
					continue
				}
				if err := sse.PatchElementTempl(pages.Info(props)); err != nil {
					l.ErrorContext(ctx, "tick: patch", "err", err)
					continue
				}
				lastProps = &props
			}
		}
	})
}

func getEducatorViewSchedule(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Schedule(pages.ScheduleProps{}).Render(ctx, w)
	}
}

func getEducatorViewScheduleStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
	periodReadModel *periodEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return sseView(l, "educator view schedule stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		username := chi.URLParam(r, "username")
		user := currentUser(r)

		render := func() (pages.ScheduleProps, error) {
			educator, err := educatorReadModel.GetByUsername(ctx, username)
			if err != nil {
				return pages.ScheduleProps{}, fmt.Errorf("get educator: %w", err)
			}

			periods, err := periodReadModel.ListPeriodsForEducator(ctx, educator.ID)
			if err != nil {
				return pages.ScheduleProps{}, fmt.Errorf("list periods: %w", err)
			}
			periodViews := scheduleDTO.NewSchedulePeriodViews(periods...)
			for index, periodView := range periodViews {
				periodModel, err := periodReadModel.GetWithIDs(ctx, periodView.Period.ID)
				if err != nil {
					l.ErrorContext(ctx, "tick: get period", "err", err, "periodID", periodView.Period.ID)
					continue
				}
				students, err := studentReadModel.ListByIDs(ctx, periodModel.StudentIDs)
				if err != nil {
					l.ErrorContext(ctx, "tick: get students", "err", err, "periodID", periodView.Period.ID)
					continue
				}
				periodViews[index].Students = studentDTO.NewViews(students)
			}

			toastList, err := listToasts(ctx, vs, user.ID)
			if err != nil {
				return pages.ScheduleProps{}, fmt.Errorf("list toasts: %w", err)
			}

			return pages.ScheduleProps{
				Toasts:   toastList,
				Educator: educator.Educator,
				Periods:  periodViews,
			}, nil
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastProps *pages.ScheduleProps

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := pruneToasts(ctx, vs, user.ID); err != nil {
					l.ErrorContext(ctx, "tick: prune toasts", "err", err)
				}
				props, err := render()
				if err != nil {
					if errors.Is(err, errs.ErrNotFound) {
						return err
					}
					l.ErrorContext(ctx, "tick: render", "err", err)
					continue
				}
				if lastProps != nil && reflect.DeepEqual(*lastProps, props) {
					continue
				}
				if err := sse.PatchElementTempl(pages.Schedule(props)); err != nil {
					l.ErrorContext(ctx, "tick: patch", "err", err)
					continue
				}
				lastProps = &props
			}
		}
	})
}

func getEducatorViewCaseload(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.Caseload(pages.CaseloadProps{}).Render(ctx, w)
	}
}

func getEducatorViewCaseloadStream(
	l *slog.Logger,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return sseView(l, "educator view caseload stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		username := chi.URLParam(r, "username")
		user := currentUser(r)

		render := func() (pages.CaseloadProps, error) {
			educator, err := educatorReadModel.GetByUsername(ctx, username)
			if err != nil {
				return pages.CaseloadProps{}, fmt.Errorf("get educator: %w", err)
			}
			caseManager, err := educatorReadModel.GetByUsernameWithCaseload(ctx, educator.Username)
			if err != nil {
				return pages.CaseloadProps{}, fmt.Errorf("get caseload: %w", err)
			}
			toastList, err := listToasts(ctx, vs, user.ID)
			if err != nil {
				return pages.CaseloadProps{}, fmt.Errorf("list toasts: %w", err)
			}
			return pages.CaseloadProps{
				Toasts:   toastList,
				Educator: educator.Educator,
				Caseload: studentDTO.NewViews(caseManager.Caseload),
			}, nil
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastProps *pages.CaseloadProps

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := pruneToasts(ctx, vs, user.ID); err != nil {
					l.ErrorContext(ctx, "tick: prune toasts", "err", err)
				}
				props, err := render()
				if err != nil {
					if errors.Is(err, errs.ErrNotFound) {
						return err
					}
					l.ErrorContext(ctx, "tick: render", "err", err)
					continue
				}
				if lastProps != nil && reflect.DeepEqual(*lastProps, props) {
					continue
				}
				if err := sse.PatchElementTempl(pages.Caseload(props)); err != nil {
					l.ErrorContext(ctx, "tick: patch", "err", err)
					continue
				}
				lastProps = &props
			}
		}
	})
}

func getEducatorCreate(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := pages.EducatorFormPageProps{
			FormType: sharedmodels.FormTypeCreate,
		}
		_ = pages.EducatorFormPage(view).Render(ctx, w)
	}
}

func getEducatorCreateStream(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return sseView(l, "educator create stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		user := currentUser(r)

		key, err := viewstore.Key(user.ID, sharedmodels.FormTypeCreate.Word(), "educators", "")
		if err != nil {
			return fmt.Errorf("view key: %w", err)
		}
		watcher, err := vs.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			return fmt.Errorf("watch: %w", err)
		}
		defer watcher.Stop()

		props, ok, err := viewstore.GetState[pages.EducatorFormPageProps](ctx, vs, key)
		if err != nil {
			return fmt.Errorf("get state: %w", err)
		}
		if !ok {
			props = pages.EducatorFormPageProps{
				FormType: sharedmodels.FormTypeCreate,
			}
			if err := viewstore.PutState(ctx, vs, key, props); err != nil {
				return fmt.Errorf("seed state: %w", err)
			}
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case entry, ok := <-watcher.Updates():
				if !ok {
					return nil
				}
				if err := entry.JSON(&props); err != nil {
					return fmt.Errorf("decode watcher entry: %w", err)
				}
				if err := sse.PatchElementTempl(pages.EducatorFormPage(props)); err != nil {
					return fmt.Errorf("render form: %w", err)
				}
			}
		}
	})
}

func postEducatorCreate(
	l *slog.Logger,
	saver eventstore.Saver,
	vs viewstore.Store,
) http.HandlerFunc {
	signals := &pages.EducatorFormPageProps{}
	return httpCommand(l, "educator create", signals, func(ctx context.Context, r *http.Request, signals *pages.EducatorFormPageProps) error {
		user := currentUser(r)
		roleString := make([]string, len(signals.Educator.Roles))
		for i, role := range signals.Educator.Roles {
			roleString[i] = role.StringLong()
		}
		educator := events.EducatorState{
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Roles:      roleString,
		}
		cmd := events.CreateEducatorCommand{
			EducatorState: educator,
			Metadata:      eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		_, err := events.CreateEducatorCommandHandler(ctx, cmd, saver)
		if err != nil {
			_ = pushToast(ctx, vs, user.ID, toastError(err.Error()))
			return nil
		}
		_ = pushToast(ctx, vs, user.ID, toastSuccess("educator created"))
		return nil
	})
}

func getEducatorEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := pages.EducatorFormPageProps{
			FormType: sharedmodels.FormTypeEdit,
		}
		_ = pages.EducatorFormPage(view).Render(ctx, w)
	}
}

func getEducatorEditStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *events.ReadModel,
) http.HandlerFunc {
	return sseView(l, "educator edit stream", func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error {
		username := chi.URLParam(r, "username")
		session, err := openEducatorSession(ctx, r, username, subscriber, vs, educatorReadModel, sharedmodels.FormTypeEdit.Word())
		if err != nil {
			return err
		}
		defer session.Close()

		build := func(e *models.Educator) pages.EducatorFormPageProps {
			return pages.EducatorFormPageProps{
				FormType: sharedmodels.FormTypeEdit,
				Educator: *e,
			}
		}

		props, ok, err := viewstore.GetState[pages.EducatorFormPageProps](ctx, vs, session.Key)
		if err != nil {
			return fmt.Errorf("get state %q: %w", session.Key, err)
		}
		if !ok {
			props = build(session.Educator)
			if err := viewstore.PutState(ctx, vs, session.Key, props); err != nil {
				return fmt.Errorf("seed state %q: %w", session.Key, err)
			}
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-session.Notifier.Signal():
				if err := refreshEducatorState(ctx, vs, session, username, educatorReadModel, build); err != nil {
					return fmt.Errorf("refresh %q: %w", username, err)
				}
			case entry, ok := <-session.Updates:
				if !ok {
					return nil
				}
				if err := entry.JSON(&props); err != nil {
					return fmt.Errorf("decode watcher entry: %w", err)
				}
				if err := sse.PatchElementTempl(pages.EducatorFormPage(props)); err != nil {
					return fmt.Errorf("render form: %w", err)
				}
			}
		}
	})
}

func postEducatorEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	vs viewstore.Store,
) http.HandlerFunc {
	signals := &pages.EducatorFormPageProps{}
	return httpCommand(l, "educator edit", signals, func(ctx context.Context, r *http.Request, signals *pages.EducatorFormPageProps) error {
		user := currentUser(r)
		roleString := make([]string, len(signals.Educator.Roles))
		for i, role := range signals.Educator.Roles {
			roleString[i] = role.StringLong()
		}
		educator := events.EducatorState{
			ID:         signals.Educator.ID,
			GivenName:  signals.Educator.GivenName,
			ChosenName: signals.Educator.ChosenName,
			FamilyName: signals.Educator.FamilyName,
			Email:      signals.Educator.Email,
			Username:   signals.Educator.Username,
			Roles:      roleString,
		}
		cmd := events.UpdateEducatorCommand{
			EducatorState: educator,
			Metadata:      eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.UpdateEducatorCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			_ = pushToast(ctx, vs, user.ID, toastError(err.Error()))
			return nil
		}
		l.InfoContext(ctx, "educator updated", "eventID", result.EventID, "skipped", result.Skipped)
		_ = pushToast(ctx, vs, user.ID, toastSuccess("saved"))
		return nil
	})
}

func queryEducatorFormValidate(
	l *slog.Logger,
	vs viewstore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		signals := &pages.EducatorFormPageProps{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			l.ErrorContext(ctx, "qefv signals", "err", err)
			return
		}
		key, err := viewstore.Key(user.ID, signals.FormType.Word(), "educators", signals.Educator.ID)
		if err != nil {
			l.ErrorContext(ctx, "qefv key", "err", err, "form type", signals.FormType, "username", user.Username, "educator id", signals.Educator.ID)
			return
		}
		if err := viewstore.PutState(ctx, vs, key, signals); err != nil {
			l.ErrorContext(ctx, "educator create validate view store", "err", err)
		}
	}
}

func deleteEducator(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	vs viewstore.Store,
) http.HandlerFunc {
	signals := &struct{}{}
	return httpCommand(l, "delete educator", signals, func(ctx context.Context, r *http.Request, _ *struct{}) error {
		user := currentUser(r)
		educatorID := chi.URLParam(r, "id")
		if educatorID == "" {
			return nil
		}
		cmd := events.DeleteEducatorCommand{
			EducatorID: educatorID,
			Metadata:   eventstore.HTTPCommandMetadata(r, user.UserRegisteredID),
		}
		result, err := events.DeleteEducatorCommandHandler(ctx, cmd, saver, retriever)
		if err != nil {
			_ = pushToast(ctx, vs, user.ID, toastError(err.Error()))
			return nil
		}
		l.InfoContext(ctx, "educator deleted", "id", educatorID, "event", result.EventID)
		_ = pushToast(ctx, vs, user.ID, toastSuccess("educator deleted"))
		return nil
	})
}

// helpers unchanged below

func getEducator(
	ctx context.Context,
	_ *slog.Logger,
	rm *events.ReadModel,
	username string,
) (*models.Educator, error) {
	view, err := rm.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &view.Educator, nil
}

func listEducators(
	ctx context.Context,
	_ *slog.Logger,
	rm *events.ReadModel,
	sort *shareddto.TableSort,
	groupedBy string,
	filters *models.Filters,
) ([]models.Educator, error) {
	stringFilter := []string{}
	if filters != nil {
		stringFilter = models.ModelToString(filters.Roles)
	}
	if filters == nil {
		filters = models.NewFilters()
	}
	if sort == nil {
		sort = &shareddto.TableSort{
			Column:    "family_name",
			Direction: "ASC",
		}
	}
	sortCol := sort.Column
	if sort.Column == "first_l" {
		sortCol = "given_name"
	}
	return rm.List(
		ctx,
		events.WithSortColumn(sortCol),
		events.WithSortDirection(sort.Direction),
		events.WithGroupedBy(groupedBy),
		events.WithSearchFilter(filters.Search),
		events.WithRoleFilter(stringFilter),
	)
}

func listEducatorsByIDs(
	ctx context.Context,
	l *slog.Logger,
	rm *events.ReadModel,
	ids []string,
) []models.Educator {
	educators, err := rm.ListByIDs(ctx, ids)
	if err != nil {
		l.ErrorContext(ctx, "list educators by ids", "err", err)
		return []models.Educator{}
	}
	return educators
}

func createEducatorSelectView(
	ctx context.Context,
	l *slog.Logger,
	educatorReadModel *events.ReadModel,
	filters *models.Filters,
	selected []string,
) dto.SelectView {
	educators, err := listEducators(ctx, l, educatorReadModel, nil, "", filters)
	if err != nil {
		l.ErrorContext(ctx, "cesv list educators", "err", err)
		return dto.SelectView{}
	}
	return dto.NewSelectView(nil, educators, selected...)
}

func roleFilterButtons(roleFilter []string) uifilters.FilterButtonsProps {
	buttons := make([]uifilters.FilterButtonProp, len(sharedmodels.EducatorRoleList))
	signalKey := "filter.role"
	for i, r := range sharedmodels.EducatorRoleList {
		active := slices.Contains(roleFilter, r.StringLong())
		buttons[i] = uifilters.FilterButtonProp{
			Label:       r.String(),
			SignalKey:   signalKey,
			SignalValue: r.StringLong(),
			Active:      active,
		}
	}
	return uifilters.FilterButtonsProps{
		Label:   "role",
		Buttons: buttons,
	}
}
