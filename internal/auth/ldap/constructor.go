package ldap

import (
	"ez2boot/internal/db"
	"log/slog"
)

func NewService(ldapRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   ldapRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
