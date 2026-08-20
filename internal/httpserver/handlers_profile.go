package httpserver

import (
	"context"
	"errors"
	"net/http"

	"seek/internal/appdb"
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
	r.Post("/profile/edit", s.postProfileEdit)
}

// GET request to /profile
func (s Server) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.profileUser(ctx, currentUser(r))
	s.Logger.Debug("test", "U", user.Email)
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
func (s Server) postProfileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.profileUser(ctx, currentUser(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	signals := &struct {
		Profile models.Profile `json:"profile"`
	}{}
	datastar.ReadSignals(r, signals)
	command := events.UpdateProfileBioCommand{
		User:     user,
		Bio:      signals.Profile.Bio,
		Metadata: eventstore.HTTPCommandMetadata(r, user.ID),
	}
	if err := events.UpdateProfileBioCommandHandler(
		ctx,
		command,
		s.EventSaver,
		s.EventRetriever,
		s.PIIKeys,
	); err != nil {
		s.Logger.ErrorContext(ctx, "profile update", "err", err)
	}
	sse := newSSE(w, r)
	sse.Redirect("/profile")
}

// refreshes profile view state in kv store when there is an update for the SSE stream
func (s Server) refreshProfileViewState(ctx context.Context, key string, current um.User) error {
	user, err := s.profileUser(ctx, current)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, key, profileViewState{User: user})
}

// helper to get profile for the user
func (s Server) profileUser(ctx context.Context, current um.User) (um.User, error) {
	user, err := s.ReadModels.Profiles.GetUserProfileByID(ctx, current.UserRegisteredID)
	if err != nil {
		if errors.Is(err, appdb.ErrNoRows) {
			s.Logger.DebugContext(ctx, "returning current user due to no profile entry in db")
			return current, nil
		}
		return um.User{}, err
	}
	return user, nil
}
