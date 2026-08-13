package httpserver

import (
	"log/slog"
	"net/http"

	"seek/internal/eventstore"
	"seek/internal/seed"
	"seek/internal/ui/core/corepages"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) coreRoutes(r chi.Router) {
	r.Get("/", getIndex())
	r.Post("/theme", postTheme(s.Logger))
	r.Get("/hub", getHub())
	r.Get("/components", getComponents())
	r.Get("/seed", getSeed())
	r.Post("/seed/educators", postSeedEducators(s.Logger, s.EventSaver))
	r.Post("/seed/students", postSeedStudents(s.Logger, s.EventSaver))
	r.Post("/seed/periods", postSeedPeriods(s.Logger, s.EventSaver))
}

// GET request to "/"
// serves as the landing page when not logged in
// dashboard when logged in
func getIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = corepages.Index().Render(ctx, w)
	}
}

// POST request to "/theme"
// nothing yet
// TODO save it to user db
func postTheme(
	l *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signals := &struct {
			Theme bool `json:"theme"`
		}{}

		err := datastar.ReadSignals(r, signals)
		if err != nil {
			l.ErrorContext(ctx, "theme signals", "err", err)
		}
		l.Debug("theme", "theme", signals.Theme)
	}
}

// GET request to "/hub"
// special rights resource hub
func getHub() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = corepages.Hub().Render(ctx, w)
	}
}

// GET request to "/components"
// development tool for viewing components and testing
func getComponents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = corepages.Components().Render(ctx, w)
	}
}

// GET request to "/seed"
// development tool for seeding data after resetting the event store
func getSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = corepages.Seed().Render(ctx, w)
	}
}

// POST request to "/seed/educators"
// seeds educator data
func postSeedEducators(
	l *slog.Logger,
	saver eventstore.Saver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, err := seed.SeedEducators(ctx, saver)
		if err != nil {
			l.ErrorContext(ctx, "seed educators", "err", err)
		}
	}
}

// POST request to "/seed/students"
// seeds student data
func postSeedStudents(
	l *slog.Logger,
	saver eventstore.Saver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, err := seed.SeedStudents(ctx, saver)
		if err != nil {
			l.ErrorContext(ctx, "seed students", "err", err)
		}
	}
}

// POST request to "/seed/educators"
// seeds period data
func postSeedPeriods(
	l *slog.Logger,
	saver eventstore.Saver,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, err := seed.SeedPeriods(ctx, saver)
		if err != nil {
			l.ErrorContext(ctx, "seed periods", "err", err)
		}
	}
}
