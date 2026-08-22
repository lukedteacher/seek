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
	r.Get("/profile", s.getProfile)
	r.Get("/profile/stream", s.getProfileStream)
	r.Get("/profile/edit", s.getEdit)
	r.Post("/profile/edit", postProfileEdit(s.Logger, s.EventSaver, s.EventRetriever, s.PIIKeys, *s.ReadModels.Profiles))
}

// GET request to /profile
func (s Server) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := profileUser(ctx, s.Logger, currentUser(r), *s.ReadModels.Profiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.Profile(user, nil).Render(ctx, w)
}

func (s Server) getProfileStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)
	key := viewstore.ProfileKey(s.sessionID(r), user.UserRegisteredID)

	notifier := NewDedupeNotifier()
	// subscribes to the channel which publishes changes to this profile
	sub, err := s.Subscriber.Subscribe(ctx, events.Channel(user.UserRegisteredID), func(context.Context, []byte) {
		notifier.Notify()
	})
	if err != nil {
		s.Logger.ErrorContext(ctx, "get profile stream subscribe", "err", err)
		return
	}
	defer sub.Close()

	watcher, err := s.ViewStore.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		s.Logger.ErrorContext(ctx, "get profile stream watcher", "err", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshProfileViewState(ctx, key, user); err != nil {
				s.Logger.ErrorContext(ctx, "get profile stream refresh", "err", err)
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var state profileViewState
			if err := entry.JSON(&state); err != nil {
				s.Logger.ErrorContext(ctx, "get profile stream json read", "err", err)
				return
			}
			sse.PatchElementTempl(pages.Profile(user, nil))
		}
	}
}

// GET request to /profile/edit
func (s Server) getEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = pages.EditProfile(nil).Render(ctx, w)
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
func (s Server) refreshProfileViewState(ctx context.Context, key string, current um.User) error {
	user, err := profileUser(ctx, s.Logger, current, *s.ReadModels.Profiles)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, key, profileViewState{User: user})
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
