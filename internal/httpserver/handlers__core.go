package httpserver

import (
	"net/http"

	"seek/internal/views/blocks/sidebar"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) coreRoutes(r chi.Router) {
	r.Get("/", s.index)
	r.Get("/stream", s.getIndexSSE)
	r.Post("/sidebar", s.postSidebarToggle)
	r.Get("/components", s.components)
	r.Post("/sort", s.sort)
	r.Get("/seed", s.seedData)
}

func (s Server) index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	view, ok, err := viewstore.GetState[sidebar.SidebarView](ctx, s.ViewStore, user.ID+".view")
	if err != nil {
		println(err.Error())
	}
	if !ok {
		println("no value found")
	}
	_ = pages.Index(user, view).Render(r.Context(), w)
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
		println(err.Error())
		return
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-watcher.Updates(): // triggers when the view state publishes to kv store
			if !ok {
				return
			}
			view := &sidebar.SidebarView{}
			if err := entry.JSON(&view); err != nil {
				println(err.Error())
				return
			}
			sse.PatchElementTempl(pages.Index(user, *view))
		}
	}
}

// TODO make this saved?
func (s Server) postSidebarToggle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &sidebar.SidebarView{}
	_ = datastar.ReadSignals(r, signals)
	user := currentUser(r)

	err := viewstore.PutState(ctx, s.ViewStore, user.ID+".view", signals)
	if err != nil {
		println(err.Error())
	}
}

func (s Server) components(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := currentUser(r)
	_ = pages.Components(user).Render(ctx, w)
}

func (s Server) sort(w http.ResponseWriter, r *http.Request) {
	signals := &struct {
		Table Table `json:"table"`
	}{}
	datastar.ReadSignals(r, signals)
	println(signals.Table.FirstName)
}

type Table struct {
	FirstName   bool `json:"first_name"`
	ChosenName  bool `json:"chosen_name"`
	LastName    bool `json:"last_name"`
	Grade       bool `json:"grade"`
	Homeroom    bool `json:"homeroom"`
	CaseManager bool `json:"case_manager"`
}
