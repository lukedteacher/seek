package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"seek/internal/appcore"
	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	profile "seek/internal/features/profiles/events"
	"seek/internal/features/users/models"
	"seek/internal/resources"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
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
		s.iepServiceRoutes(r)
		s.periodRoutes(r)
		s.studentRoutes(r)
		s.profileRoutes(r)
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
			// l.InfoContext(ctx, "add path to context", "uri", r.RequestURI, "path", r.URL.Path)
			// strips the url of "/stream" and any parameters
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
		// if there is no user logged in
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
