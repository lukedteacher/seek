package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	pse "seek/internal/features/periods_students/events"
	"seek/internal/features/schedule"
	student "seek/internal/features/students/events"
	"seek/internal/features/teacher"
	"seek/internal/resources"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
)

type MessageSubscriber interface {
	Subscribe(ctx context.Context, subject string, handle func(context.Context, []byte)) (eventstore.MessageSubscription, error)
}

type Server struct {
	// Accounts       AccountCommands
	// Sessions       SessionManager
	// AuthUsers      AuthUserReader
	Periods        period.PeriodReadModelReader
	Schedules      schedule.ScheduleReadModelReader
	Students       student.StudentReadModelReader
	Teachers       teacher.TeacherReadModelReader
	PeriodStudents pse.PeriodStudentReadModelReader
	EventSaver     eventstore.Saver
	EventRetriever eventstore.Retriever
	// ProfileStorage profile.ObjectStore
	Subscriber  MessageSubscriber
	ViewStore   viewstore.Store
	Development bool
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

	// s.authRoutes(r)

	r.Group(func(r chi.Router) {
		// 	r.Use(s.requireVerifiedEmail)
		s.coreRoutes(r)
		s.periodRoutes(r)
		s.scheduleRoutes(r)
		s.studentRoutes(r)
		s.teacherRoutes(r)
		// 	s.profileRoutes(r)
	})

	return r
}
