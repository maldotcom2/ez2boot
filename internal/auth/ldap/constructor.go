package ldap

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/encryption"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(ldapService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service:  ldapService,
		Searcher: ldapService, // for testing
		Logger:   logger,
	}
}

func NewService(ldapRepo *Repository, userService *user.Service, audit *audit.Service, encryptor *encryption.AESGCMEncryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:        ldapRepo,
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
