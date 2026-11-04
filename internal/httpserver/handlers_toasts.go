package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
)

func (s Server) toastRoutes(r chi.Router) {
	r.Delete("/toasts/{id}", deleteToast(s.Logger, s.ViewStore))
}

func toastSuccess(msg string) sharedmodels.Toast { return newToast("success", msg) }
func toastError(msg string) sharedmodels.Toast   { return newToast("error", msg) }
func toastInfo(msg string) sharedmodels.Toast    { return newToast("info", msg) }

func newToast(kind, msg string) sharedmodels.Toast {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return sharedmodels.Toast{
		ID:       hex.EncodeToString(b[:]),
		Kind:     kind,
		Message:  msg,
		SentAt:   time.Now(),
		Duration: sharedmodels.ToastLifetime,
	}
}

func toastKey(userID string) (string, error) {
	return viewstore.Key(userID, "list", "toasts", "")
}

func listToasts(ctx context.Context, vs viewstore.Store, userID string) ([]sharedmodels.Toast, error) {
	k, err := toastKey(userID)
	if err != nil {
		return nil, err
	}
	current, _, err := viewstore.GetState[[]sharedmodels.Toast](ctx, vs, k)
	if err != nil {
		return nil, err
	}
	return liveToasts(current, time.Now()), nil
}

func pushToast(ctx context.Context, vs viewstore.Store, userID string, t sharedmodels.Toast) error {
	k, err := toastKey(userID)
	if err != nil {
		return err
	}
	current, _, err := viewstore.GetState[[]sharedmodels.Toast](ctx, vs, k)
	if err != nil {
		return err
	}
	current = liveToasts(current, time.Now())
	current = append(current, t)
	if len(current) > sharedmodels.ToastMax {
		current = current[len(current)-sharedmodels.ToastMax:]
	}
	return viewstore.PutState(ctx, vs, k, current)
}

func pruneToasts(ctx context.Context, vs viewstore.Store, userID string) error {
	k, err := toastKey(userID)
	if err != nil {
		return err
	}
	current, _, err := viewstore.GetState[[]sharedmodels.Toast](ctx, vs, k)
	if err != nil {
		return err
	}
	next := liveToasts(current, time.Now())
	if len(next) == len(current) {
		return nil
	}
	return viewstore.PutState(ctx, vs, k, next)
}

func liveToasts(list []sharedmodels.Toast, now time.Time) []sharedmodels.Toast {
	out := make([]sharedmodels.Toast, 0, len(list))
	for _, t := range list {
		if t.Live(now) {
			out = append(out, t)
		}
	}
	return out
}

func deleteToast(l *slog.Logger, vs viewstore.Store) http.HandlerFunc {
	signals := &struct{}{}
	return httpCommand(l, "delete toast", signals, func(ctx context.Context, r *http.Request, _ *struct{}) error {
		user := currentUser(r)
		toastID := chi.URLParam(r, "id")
		if toastID == "" {
			return nil
		}
		k, err := toastKey(user.ID)
		if err != nil {
			return fmt.Errorf("toast key: %w", err)
		}
		current, _, err := viewstore.GetState[[]sharedmodels.Toast](ctx, vs, k)
		if err != nil {
			return fmt.Errorf("toast read: %w", err)
		}
		out := current[:0]
		for _, t := range current {
			if t.ID != toastID {
				out = append(out, t)
			}
		}
		if err := viewstore.PutState(ctx, vs, k, out); err != nil {
			return fmt.Errorf("toast write: %w", err)
		}
		return nil
	})
}
