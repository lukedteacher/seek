package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"seek/internal/auth"
	"seek/internal/domain/models"
	"seek/internal/eventstore"
	iepService "seek/internal/features/iep_services/events"
	period "seek/internal/features/periods/events"
	periodSchedule "seek/internal/features/periods_schedules/events"
	periodStudent "seek/internal/features/periods_students/events"
	profile "seek/internal/features/profiles/events"
	schedule "seek/internal/features/schedules/events"
	student "seek/internal/features/students/events"
	teacher "seek/internal/features/teachers/events"
	"seek/internal/resources"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const userKey contextKey = "user"

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

type VerificationStore interface {
	auth.PasswordResetReader
}

type Server struct {
	Sessions            SessionManager
	AuthUsers           AuthUserReader
	PIIKeys             auth.SubjectPiiKeyPort
	PasswordCredentials auth.PasswordCredentialReader
	Verifications       VerificationStore
	IEPServices         iepService.IEPServiceReadModelReader
	Periods             period.PeriodReadModelReader
	Profiles            ProfileReader
	Schedules           schedule.ScheduleReadModelReader
	Students            student.StudentReadModelReader
	Teachers            teacher.TeacherReadModelReader
	PeriodsSchedules    periodSchedule.PeriodScheduleReadModelReader
	PeriodsStudents     periodStudent.PeriodStudentReadModelReader
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

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})
	if s.Development {
		setupReload(r)
	}
	r.Handle("/static/*", resources.Handler())
	r.Group(func(r chi.Router) {
		r.Use(s.addUserInfoToContext)
		s.authRoutes(r)
		s.coreRoutes(r)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.requireVerifiedEmail)
		s.iepServiceRoutes(r)
		s.periodRoutes(r)
		s.scheduleRoutes(r)
		s.studentRoutes(r)
		s.teacherRoutes(r)
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userKey, user)))
	})
}

func (s Server) requireVerifiedEmail(next http.Handler) http.Handler {
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
		if !user.EmailVerified {
			http.Redirect(w, r, "/register/"+user.ID+"/validate-email", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userKey, user)))
	})
}

func currentUser(r *http.Request) models.User {
	user, _ := r.Context().Value(userKey).(models.User)
	return user
}

func (s Server) sessionID(r *http.Request) string {
	cookie, err := r.Cookie(s.Sessions.SessionCookieName())
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return currentUser(r).UserRegisteredID
}
