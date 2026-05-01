package http

import (
	"net/http"

	"task-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(actor *storage.Actor) *Server {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(RequestLogger)

	h := NewHandler(actor)

	// Static assets
	r.Get("/", h.Index)

	// HTMX partials
	r.Post("/tasks/create", h.CreateTask)
	r.Post("/tasks/status", h.UpdateStatus)
	r.Delete("/tasks/delete", h.DeleteTask)

	// JSON API
	r.Route("/api/tasks", func(r chi.Router) {
		r.Get("/", h.APIGetTasks)
		r.Post("/", h.APICreateTask)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.APIGetTask)
			r.Put("/", h.APIUpdateTask)
			r.Delete("/", h.APIDeleteTask)
		})
	})

	// Serve HTMX CDN locally as fallback (not needed with CDN, but nice to have)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return &Server{Router: r}
}

func chiRouteContext(r *http.Request) chiContext {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		return routeCtx
	}
	return nil
}

// chiContext interface matches chi's Context for URLParam extraction.
type chiContext interface {
	URLParam(key string) string
}
