package auth

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
)

type Service struct {
	UserService *user.Service
	LdapService *ldap.Service
	Config      *config.Config
	Audit       *audit.Service
	Logger      *slog.Logger
}

type Handler struct {
	Service *Service
	Config  *config.Config
	Logger  *slog.Logger
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MFARequiredResponse struct {
	MFARequired bool `json:"mfa_required"`
}
