package app

import (
	"context"
	"errors"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/shared"
	"ez2boot/internal/worker"
	"fmt"
	"log/slog"
	"net/http"
)

func NewApp(version string, buildDate string, cfg *config.Config, repo *db.Repository, logger *slog.Logger) (http.Handler, *Services, *worker.Worker, error) {
	// Fail fast if config is invalid
	if err := validateProviderConfig(cfg); err != nil {
		return nil, nil, nil, err
	}

	mw, wkr, handlers, services, err := InitServices(version, buildDate, cfg, repo, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	// Check if there are existing users and set runtime mode
	hasUsers, err := hasUsers(repo)
	if err != nil {
		return nil, nil, nil, err
	}

	cfg.SetupMode = !hasUsers

	// Initialise OIDC provider if configured
	if err := services.OidcService.InitProvider(context.Background()); err != nil {
		if !errors.Is(err, shared.ErrOIDCConfigNotFound) {
			return nil, nil, nil, fmt.Errorf("failed to init OIDC provider: %w", err)
		}
	}

	router := BuildRouter(cfg, mw, handlers)

	return router, services, wkr, nil
}
