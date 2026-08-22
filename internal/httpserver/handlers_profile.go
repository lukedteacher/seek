package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"seek/internal/appdb"
	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/features/profiles/events"
	"seek/internal/features/profiles/models"
	"seek/internal/features/profiles/pages"
	um "seek/internal/features/users/models"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

type profileViewState struct {
	User um.User `json:"user"`
}

func (s Server) profileRoutes(r chi.Router) {
	r.Get("/profile", getProfile(s.Logger, *s.ReadModels.Profiles))
	r.Get("/profile/stream", getProfileStream(s.Logger, s.Subscriber, s.ViewStore, s.sessionID, *s.ReadModels.Profiles))
	r.Get("/profile/edit", getEdit(s.Logger))
	r.Post("/profile/edit", postProfileEdit(s.Logger, s.EventSaver, s.EventRetriever, s.PIIKeys, *s.ReadModels.Profiles))
}

// GET request to /profile
func getProfile(
	l *slog.Logger,
	profileReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := profileUser(ctx, l, currentUser(r), profileReadModel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = pages.Profile(user, nil).Render(ctx, w)
	}
}

func getProfileStream(
	l *slog.Logger,
	subscriber MessageSubscriber,
	vs viewstore.Store,
	sessionID func(r *http.Request) string,
	profileReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := currentUser(r)
		sse := newSSE(w, r)
		key := viewstore.ProfileKey(sessionID(r), user.UserRegisteredID)

		notifier := NewDedupeNotifier()
		// subscribes to the channel which publishes changes to this profile
		sub, err := subscriber.Subscribe(ctx, events.Channel(user.UserRegisteredID), func(context.Context, []byte) {
			notifier.Notify()
		})
		if err != nil {
			l.ErrorContext(ctx, "get profile stream subscribe", "err", err)
			return
		}
		defer sub.Close()

		watcher, err := vs.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
		if err != nil {
			l.ErrorContext(ctx, "get profile stream watcher", "err", err)
			return
		}
		defer watcher.Stop()

		if err := refreshProfileViewState(ctx, l, vs, key, user, profileReadModel); err != nil {
			l.ErrorContext(ctx, "get profile stream refresh", "err", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-notifier.Signal():
				if err := refreshProfileViewState(ctx, l, vs, key, user, profileReadModel); err != nil {
					l.ErrorContext(ctx, "get profile stream refresh", "err", err)
					return
				}
			case entry, ok := <-watcher.Updates():
				if !ok {
					return
				}
				var state profileViewState
				if err := entry.JSON(&state); err != nil {
					l.ErrorContext(ctx, "get profile stream json read", "err", err)
					return
				}
				sse.PatchElementTempl(pages.Profile(user, nil))
			}
		}
	}
}

// GET request to /profile/edit
func getEdit(
	_ *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = pages.EditProfile(nil).Render(ctx, w)
	}
}

// POST request to /profile/edit
func postProfileEdit(
	l *slog.Logger,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	piiKeys auth.SubjectPiiKeyPort,
	profileReadModel events.ReadModel,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()
		// user, err := profileUser(ctx, l, currentUser(r), profileReadModel)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		signals := &struct {
			Profile models.Profile `json:"profile"`
		}{}
		datastar.ReadSignals(r, signals)
		l.Debug("post profile edit", "a", signals.Profile.Avatar)
		sse := newSSE(w, r)
		sse.Redirect("/profile")
	}
}

// refreshes profile view state in kv store when there is an update for the SSE stream
func refreshProfileViewState(
	ctx context.Context,
	l *slog.Logger,
	vs viewstore.Store,
	key string,
	current um.User,
	profileReadModel events.ReadModel,
) error {
	user, err := profileUser(ctx, l, current, profileReadModel)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, vs, key, profileViewState{User: user})
}

// helper to get profile for the user
func profileUser(
	ctx context.Context,
	l *slog.Logger,
	current um.User,
	profileReadModel events.ReadModel,
) (um.User, error) {
	user, err := profileReadModel.GetUserProfileByID(ctx, current.UserRegisteredID)
	if err != nil {
		if errors.Is(err, appdb.ErrNoRows) {
			l.DebugContext(ctx, "returning current user due to no profile entry in db")
			return current, nil
		}
		return um.User{}, err
	}
	return user, nil
}
