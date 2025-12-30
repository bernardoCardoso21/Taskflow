package http

import (
	"log"
	"net/http"

	"TaskFlow/internal/config"
	"TaskFlow/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Deps struct {
	Config     config.Config
	AuthSvc    *service.AuthService
	ProjectSvc *service.ProjectService
	TaskSvc    *service.TaskService
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(RecovererJSON)
	r.Use(middleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://localhost:8081/taskflow/swaggerui", "http://localhost:8081/*", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, 200, map[string]any{"status": "ok"})
	})

	authH := NewAuthHandler(d.AuthSvc)
	projH := NewProjectHandler(d.ProjectSvc)
	taskH := NewTaskHandler(d.TaskSvc)

	r.Route("/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(AuthJWT(d.Config.JWTSecret))

			// auth/me
			r.Get("/auth/me", authH.Me)

			// projects
			r.Route("/projects", func(r chi.Router) {
				r.Post("/", projH.Create)
				r.Get("/", projH.List)
				r.Get("/{id}", projH.Get)
				r.Patch("/{id}", projH.Update)
				r.Delete("/{id}", projH.Delete)

				// tasks under a project
				r.Post("/{projectId}/tasks", taskH.Create)
			})

			// tasks
			r.Get("/tasks", taskH.List)
			r.Get("/tasks/{id}", taskH.Get)
			r.Patch("/tasks/{id}", taskH.Update)
			r.Delete("/tasks/{id}", taskH.Delete)
		})
	})

	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s", method, route)
		return nil
	})

	return r
}
