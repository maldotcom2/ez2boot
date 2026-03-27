package ldap

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(ldapService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: ldapService,
		Logger:  logger,
	}
}

func NewService(ldapRepo *Repository, userService *user.Service, audit *audit.Service, encryptor Encryptor, logger *slog.Logger) *Service {
	s := &Service{
		Repo:        ldapRepo,
		UserService: userService,
		Audit:       audit,
		Encryptor:   encryptor,
		Logger:      logger,
	}
	// Searcher defaults to the service itself. Replace with a stub in tests
	// to avoid real LDAP connections.
	s.Searcher = s
	return s
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
