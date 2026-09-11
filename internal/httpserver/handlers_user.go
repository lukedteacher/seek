package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"seek/internal/eventstore"
	educatorEvents "seek/internal/features/educators/events"
	periodEvents "seek/internal/features/periods/events"
	scheduleDTO "seek/internal/features/schedules/dto"
	studentDTO "seek/internal/features/students/dto"
	studentEvents "seek/internal/features/students/events"
	bookmarkEvents "seek/internal/features/users/bookmarks/events"
	userModels "seek/internal/features/users/models"
	"seek/internal/ui/core/corepages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
)

func (s *Server) userRoutes(r chi.Router) {
	r.Get("/dashboard", getDashboard(s.Logger))
	r.Get("/dashboard/stream", getDashboardStream(s.Logger, s.Subscriber, s.ViewStore, s.ReadModels.Educators, s.ReadModels.Periods, s.ReadModels.Students))
	r.Post("/users/{uid}/bookmarks/students/{sid}", postStudentBookmarks(s.Logger, s.EventSaver, s.EventRetriever, s.ReadModels.Students))
}

// GET request to "/dashboard"
// dashboard when logged in
func getDashboard(
	l *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		view := corepages.DashboardView{}
		_ = corepages.Dashboard(view).Render(ctx, w)
	}
}

func getDashboardStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	educatorReadModel *educatorEvents.ReadModel,
	periodReadModel *periodEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)

		// subscribes to the channel which publishes changes to the underlying user model
		userNotifier := NewDedupeNotifier()
		subKey := "users." + user.ID
		sub, err := subscriber.Subscribe(ctx, subKey, func(context.Context, []byte) {
			userNotifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "dashboard stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		educatorID, err := refreshDashboardViewState(ctx, l, vs, user, educatorReadModel, periodReadModel, studentReadModel)
		if err != nil {
			l.ErrorContext(ctx, "dashboard stream refresh", "err", err)
			return
		}

		// subscribes to the channel which publishes changes to the underlying user model
		educatorNotifier := NewDedupeNotifier()
		educatorSubKey := "educators." + educatorID
		educatorSub, err := subscriber.Subscribe(ctx, educatorSubKey, func(context.Context, []byte) {
			educatorNotifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "dashboard stream subscribe", "err", err)
			return
		}
		defer educatorSub.Close()

		// watches the key value stream for ephemeral changes
		// lasts 5m
		key := subKey + ".dashboard"
		watcher, err := vs.Watch(
			ctx,
			key,
			viewstore.WatchOptions{
				IgnoreDeletes: true,
			},
		)
		if err != nil {
			l.ErrorContext(ctx, "dashboard stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-userNotifier.Signal(): // triggers when the read model publishes
				if _, err := refreshDashboardViewState(ctx, l, vs, user, educatorReadModel, periodReadModel, studentReadModel); err != nil {
					l.ErrorContext(ctx, "dashboard stream refresh in select", "err", err)
					return
				}
			case <-educatorNotifier.Signal(): // triggers when the read model publishes
				if _, err := refreshDashboardViewState(ctx, l, vs, user, educatorReadModel, periodReadModel, studentReadModel); err != nil {
					l.ErrorContext(ctx, "dashboard stream refresh in select", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
				if !ok {
					return
				}
				view := &corepages.DashboardView{}
				if err := entry.JSON(view); err != nil {
					l.ErrorContext(ctx, "dashboard stream json read", "err", err)
					return
				}
				sse.PatchElementTempl(corepages.Dashboard(*view))
			}
		}
	}
}

func postStudentBookmarks(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	studentReadModel *studentEvents.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := chi.URLParam(r, "uid")
		studentID := chi.URLParam(r, "sid")
		student, _ := studentReadModel.GetStudentBookmark(ctx, userID, studentID)
		if student == nil {
			cmd := bookmarkEvents.AddStudentBookmarkCommand{
				UserID:    userID,
				StudentID: studentID,
				Metadata:  eventstore.HTTPCommandMetadata(r, userID),
			}
			result, err := bookmarkEvents.AddStudentBookmarkCommandHandler(
				ctx,
				cmd,
				saver,
				retriever,
			)
			if err != nil {
				l.ErrorContext(ctx, "post student bookmarks", "err", err, "uid", userID, "sid", studentID)
				return
			}
			l.InfoContext(ctx, "post student bookmarks", "skipped", result.Skipped, "uid", userID, "sid", studentID)
		} else {
			cmd := bookmarkEvents.RemoveStudentBookmarkCommand{
				UserID:    userID,
				StudentID: studentID,
			}
			result, err := bookmarkEvents.RemoveStudentBookmarkCommandHandler(
				ctx,
				cmd,
				saver,
				retriever,
			)
			if err != nil {
				l.ErrorContext(ctx, "post student bookmarks", "err", err, "uid", userID, "sid", studentID)
				return
			}
			l.InfoContext(ctx, "post student bookmarks", "skipped", result.Skipped, "uid", userID, "sid", studentID)
		}
	}
}

func refreshDashboardViewState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	user userModels.User,
	educatorReadModel *educatorEvents.ReadModel,
	periodReadModel *periodEvents.ReadModel,
	studentReadModel *studentEvents.ReadModel,
) (string, error) {
	educator, _ := educatorReadModel.GetByUsername(ctx, user.Username)
	var schedule scheduleDTO.PersonWithScheduleView
	if educator != nil {
		periods, _ := periodReadModel.ListPeriodsForEducator(ctx, educator.ID)
		if len(periods) > 0 {
			schedule = scheduleDTO.NewPersonScheduleView(educator.ID, educator.Person, periods, true, 1)
		}
	}
	if schedule.ID == "" {
		schedule = scheduleDTO.PersonWithScheduleView{}
	}
	studentBookmarks, err := studentReadModel.ListStudentBookmarksByUserID(ctx, user.ID)
	if err != nil {
		l.ErrorContext(ctx, "dashboard", "err", err, "uid", user.ID)
		return "", err
	}
	studentViews := make([]studentDTO.StudentView, len(studentBookmarks))
	for i, student := range studentBookmarks {
		studentViews[i] = studentDTO.NewView(&student)
	}
	view := corepages.DashboardView{
		Schedule:           schedule,
		BookmarkedStudents: studentViews,
	}
	key := "users." + user.ID + ".dashboard"
	err = viewstore.PutState(ctx, vs, key, view)
	if err != nil {
		return "", err
	}
	return educator.ID, nil
}
