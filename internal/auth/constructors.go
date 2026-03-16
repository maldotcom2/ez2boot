package auth

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(authService *Service, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		Service: authService,
		Config:  cfg,
		Logger:  logger,
	}
}

func NewService(userService *user.Service, ldapService LdapAuthenticator, cfg *config.Config, auditService *audit.Service, logger *slog.Logger) *Service {
	return &Service{
		UserService: userService,
		LdapService: ldapService,
		Config:      cfg,
		Audit:       auditService,
		Logger:      logger,
	}
}
