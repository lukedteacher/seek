package httpserver

import (
	"context"
	"errors"
	"net/http"

	"seek/internal/appdb"
	"seek/internal/features/profiles/pages"
	"seek/internal/domain/models"

	"github.com/go-chi/chi/v5"
)

type profileViewState struct {
	User models.User `json:"user"`
}

func (s Server) profileRoutes(r chi.Router) {
	r.Get("/profile", s.getView)
}

// GET request to /profile
func (s Server) getView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.profileUser(ctx, currentUser(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = pages.Profile(user, nil).Render(ctx, w)
}

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
