package httpserver

import (
	"context"
	"errors"
	"net/http"

	"seek/internal/appdb"
	"seek/internal/domain/models"
	"seek/internal/features/profiles/events"
	"seek/internal/features/profiles/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
)

type profileViewState struct {
	User models.User `json:"user"`
}

func (s Server) profileRoutes(r chi.Router) {
	r.Get("/profile", s.getProfile)
	r.Get("/profile/stream", s.getProfileStream)
	r.Get("/profile/edit", s.getEdit)
}

// GET request to /profile
func (s Server) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.profileUser(ctx, currentUser(r))
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
		println(err.Error())
		return
	}
	defer sub.Close()

	watcher, err := s.ViewStore.Watch(ctx, key, viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		println(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-notifier.Signal():
			if err := s.refreshProfileViewState(ctx, key, user); err != nil {
				_ = alert(sse, err.Error())
				return
			}
		case entry, ok := <-watcher.Updates():
			if !ok {
				return
			}
			var state profileViewState
			if err := entry.JSON(&state); err != nil {
				_ = alert(sse, err.Error())
				return
			}
			if err := sse.PatchElementTempl(pages.Profile(user, nil)); err != nil {
				return
			}
		}
	}
}

// GET request to /profile/edit
func (s Server) getEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.profileUser(ctx, currentUser(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.EditProfile(user, nil).Render(ctx, w)
}

// refreshes profile view state in kv store when there is an update for the SSE stream
func (s Server) refreshProfileViewState(ctx context.Context, key string, current models.User) error {
	user, err := s.profileUser(ctx, current)
	if err != nil {
		return err
	}
	return viewstore.PutState(ctx, s.ViewStore, key, profileViewState{User: user})
}

// helper to get profile for the user
func (s Server) profileUser(ctx context.Context, current models.User) (models.User, error) {
	user, err := s.Profiles.User(ctx, current.UserRegisteredID)
	if err != nil {
		if errors.Is(err, appdb.ErrNoRows) {
			return current, nil
		}
		return models.User{}, err
	}
	user.EmailVerified = current.EmailVerified
	return user, nil
}
