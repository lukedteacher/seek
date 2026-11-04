package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"seek/internal/appcore"
	"seek/internal/auth"
	"seek/internal/errs"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	profile "seek/internal/features/profiles/events"
	"seek/internal/features/users/models"
	"seek/internal/resources"
	"seek/internal/ui/core/corepages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

type MessageSubscriber interface {
	Subscribe(ctx context.Context, subject string, handle func(context.Context, []byte)) (eventstore.MessageSubscription, error)
}

type SessionManager interface {
	Login(ctx context.Context, emailAddress, password string) (models.User, string, error)
	Logout(ctx context.Context, token string) error
	CurrentUser(ctx context.Context, r *http.Request) (models.User, bool, error)
	SetSessionCookie(w http.ResponseWriter, token string)
	ClearSessionCookie(w http.ResponseWriter)
	SessionCookieName() string
}

type AuthUserReader interface {
	UserByIDOrRegisteredID(ctx context.Context, id string) (models.User, error)
}

type ProfileReader interface {
	User(ctx context.Context, userRegisteredID string) (models.User, error)
}

type Server struct {
	Sessions            SessionManager
	AuthUsers           AuthUserReader
	PIIKeys             auth.SubjectPiiKeyPort
	PasswordCredentials auth.PasswordCredentialReader
	ReadModels          appcore.ReadModelContainer
	EventSaver          eventstore.Saver
	EventRetriever      eventstore.Retriever
	ProfileStorage      profile.ObjectStore
	Subscriber          MessageSubscriber
	ViewStore           viewstore.Store
	Development         bool
	Logger              *slog.Logger
}

func (s Server) Routes() http.Handler {
	r := chi.NewRouter()
	// r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)
	// r.Use(middleware.Compress(5,
	// 	"text/html",
	// 	"text/css",
	// 	"text/plain",
	// 	"text/javascript",
	// 	"application/javascript",
	// 	"application/json",
	// 	"image/svg+xml",
	// ))

	if s.Development {
		setupReload(r)
	}
	r.Handle("/static/*", resources.Handler())
	r.Group(func(r chi.Router) {
		r.Use(s.addUserInfoToContext)
		r.Use(addPathToContext(s.Logger))
		s.authRoutes(r)
		s.coreRoutes(r)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.requireUserLoggedIn)
		r.Use(addPathToContext(s.Logger))
		s.educatorRoutes(r)
		s.homeroomRoutes(r)
		s.serviceRoutes(r)
		s.periodRoutes(r)
		s.studentRoutes(r)
		s.iepRoutes(r)
		s.profileRoutes(r)
		s.userRoutes(r)
		s.toastRoutes(r)
	})

	return r
}

func render(component interface {
	Render(context.Context, io.Writer) error
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = component.Render(r.Context(), w)
	}
}

// htmlHandler writes one HTML response. error → mapped to http status.
func htmlHandler(
	logger *slog.Logger,
	op string,
	fn func(ctx context.Context, r *http.Request) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := fn(ctx, r); err != nil {
			logger.ErrorContext(ctx, op, "err", err)
			switch {
			case errors.Is(err, errs.ErrNotFound):
				http.NotFound(w, r)
			case errors.Is(err, errs.ErrForbidden):
				http.Error(w, "forbidden", http.StatusForbidden)
			case errors.Is(err, errs.ErrConflict):
				http.Error(w, "conflict", http.StatusConflict)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}
	}
}

// sseView runs a GET stream. errors patch not-found/forbidden pages.
func sseView(
	logger *slog.Logger,
	op string,
	fn func(ctx context.Context, r *http.Request, sse *datastar.ServerSentEventGenerator) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sse := newSSE(w, r)
		if err := fn(ctx, r, sse); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			logger.ErrorContext(ctx, op, "err", err)
			switch {
			case errors.Is(err, errs.ErrNotFound):
				_ = sse.PatchElementTempl(corepages.NotFound())
			case errors.Is(err, errs.ErrForbidden):
				_ = sse.PatchElementTempl(corepages.Forbidden())
			}
		}
	}
}

// httpCommand reads signals, runs fn, returns 204. fn writes user-visible
// outcomes to the viewstore; its errors are infra failures, logged not surfaced.
func httpCommand[T any](
	logger *slog.Logger,
	op string,
	signals *T,
	fn func(ctx context.Context, r *http.Request, signals *T) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := datastar.ReadSignals(r, signals); err != nil {
			logger.ErrorContext(ctx, op, "read signals", "err", err)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err := fn(ctx, r, signals); err != nil {
			logger.ErrorContext(ctx, op, "err", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) addUserInfoToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _, err := s.Sessions.CurrentUser(ctx, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, models.UserKey, user)))
	})
}

func addPathToContext(
	_ *slog.Logger,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			baseURL := strings.TrimSuffix(r.URL.Path, "/stream")
			ctx = sharedmodels.WithURL(ctx, baseURL)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s Server) requireUserLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok, err := s.Sessions.CurrentUser(ctx, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, models.UserKey, user)))
	})
}

func currentUser(r *http.Request) models.User {
	user, _ := r.Context().Value(models.UserKey).(models.User)
	return user
}

func (s Server) sessionID(r *http.Request) string {
	cookie, err := r.Cookie(s.Sessions.SessionCookieName())
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return currentUser(r).UserRegisteredID
}
