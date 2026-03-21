package oidc

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(oidcService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: oidcService,
		Logger:  logger,
	}
}

func NewService(oidcRepo *Repository, userService *user.Service, audit *audit.Service, encryptor Encryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:        oidcRepo,
		UserService: userService,
		Audit:       audit,
		Encryptor:   encryptor,
		Logger:      logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
