package http

import (
	"net/http"

	"TaskFlow/internal/config"
	"TaskFlow/internal/service"

	"github.com/go-chi/chi/v5"
)

type Deps struct {
	Config     config.Config
	AuthSvc    *service.AuthService
	ProjectSvc *service.ProjectService
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(RecovererJSON)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, 200, map[string]any{"status": "ok"})
	})

	authH := NewAuthHandler(d.AuthSvc)

	projH := NewProjectHandler(d.ProjectSvc)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)

			r.Group(func(r chi.Router) {
				r.Use(AuthJWT(d.Config.JWTSecret))
				r.Get("/me", authH.Me)
			})
			r.Group(func(r chi.Router) {
				r.Use(AuthJWT(d.Config.JWTSecret))

				r.Route("/projects", func(r chi.Router) {
					r.Post("/", projH.Create)
					r.Get("/", projH.List)
					r.Get("/{id}", projH.Get)
					r.Patch("/{id}", projH.Update)
					r.Delete("/{id}", projH.Delete)
				})
			})
		})
	})

	return r
}
