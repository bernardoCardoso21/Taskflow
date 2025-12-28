package app

import (
	"net/http"

	"TaskFlow/internal/config"
	httpx "TaskFlow/internal/http"
	"TaskFlow/internal/repo/postgres"
	"TaskFlow/internal/service"
)

type App struct {
	Config config.Config
	Router http.Handler
}

func New() (*App, error) {
	cfg := config.FromEnv()

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	userRepo := postgres.NewUserRepo(db)
	projectRepo := postgres.NewProjectRepo(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	projectSvc := service.NewProjectService(projectRepo)

	router := httpx.NewRouter(httpx.Deps{
		Config:     cfg,
		AuthSvc:    authSvc,
		ProjectSvc: projectSvc,
	})

	return &App{
		Config: cfg,
		Router: router,
	}, nil
}
