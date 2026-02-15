package audit

import (
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(auditService *Service, adminChecker AdminChecker, logger *slog.Logger) *Handler {
	return &Handler{
		Service:      auditService,
		AdminChecker: adminChecker,
		Logger:       logger,
	}
}

func NewService(auditRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   auditRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
