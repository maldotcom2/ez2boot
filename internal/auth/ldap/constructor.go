package ldap

import (
	"ez2boot/internal/db"
	"ez2boot/internal/encryption"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(ldapService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: ldapService,
		Logger:  logger,
	}
}

func NewService(ldapRepo *Repository, userService *user.Service, encryptor *encryption.AESGCMEncryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:        ldapRepo,
		Encryptor:   encryptor,
		UserService: userService,
		Logger:      logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
