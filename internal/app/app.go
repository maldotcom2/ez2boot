package app

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/worker"
	"log/slog"
	"net/http"
)

func NewApp(version string, buildDate string, cfg *config.Config, repo *db.Repository, logger *slog.Logger) (http.Handler, *Services, *worker.Worker, error) {
	mw, wkr, handlers, services := InitServices(version, buildDate, cfg, repo, logger)

	// Check if there are existing users and set runtime mode
	hasUsers, err := hasUsers(repo)
	if err != nil {
		return nil, nil, nil, err
	}

	cfg.SetupMode = !hasUsers

	router := BuildRouter(cfg, mw, handlers)

	return router, services, wkr, nil
}
