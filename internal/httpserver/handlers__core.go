package httpserver

import (
	"net/http"

	"seek/internal/seed"
	"seek/internal/ui/core/corepages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) coreRoutes(r chi.Router) {
	r.Get("/", s.index)
	r.Get("/stream", s.getIndexSSE)
	r.Get("/components", s.components)
	r.Get("/seed", s.getSeed)
	r.Post("/seed/educators", s.postSeedEducators)
	r.Post("/seed/periods", s.postSeedPeriods)
	r.Post("/seed/students", s.postSeedStudents)
	r.Get("/hub", s.getHub)
	r.Post("/sort", s.sort)
}

func (s Server) index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = corepages.Index().Render(ctx, w)
}

func (s Server) getHub(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = corepages.Hub().Render(ctx, w)
}

func (s Server) getSeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = corepages.Seed().Render(ctx, w)
}

func (s Server) postSeedStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seed.SeedStudents(ctx, s.EventSaver)
}

func (s Server) postSeedEducators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seed.SeedEducators(ctx, s.EventSaver)
}

func (s Server) postSeedPeriods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seed.SeedPeriods(ctx, s.EventSaver)
}

func (s Server) getIndexSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	sse := newSSE(w, r)
	// watches the key value stream for ephemeral changes
	// lasts 5m
	watcher, err := s.ViewStore.Watch(
		ctx,
		user.ID+".view",
		viewstore.WatchOptions{
			IgnoreDeletes: true,
		},
	)
	if err != nil {
		s.Logger.ErrorContext(ctx, "get index sse watcher", "err", err)
		return
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			sse.PatchElementTempl(corepages.Index())
		}
	}
}

func (s Server) components(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = corepages.Components().Render(ctx, w)
}

func (s Server) sort(w http.ResponseWriter, r *http.Request) {
	signals := &struct {
		Table Table `json:"table"`
	}{}
	datastar.ReadSignals(r, signals)
}

type Table struct {
	GivenName   bool `json:"given_name"`
	ChosenName  bool `json:"chosen_name"`
	FamilyName  bool `json:"family_name"`
	Grade       bool `json:"grade"`
	Homeroom    bool `json:"homeroom"`
	CaseManager bool `json:"case_manager"`
}
