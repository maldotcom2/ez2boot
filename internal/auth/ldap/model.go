package ldap

import (
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo   *Repository
	Logger *slog.Logger
}

type LdapConfig struct {
	Host          string
	Port          int
	BaseDN        string
	BindDN        string
	BindPassword  string
	UseSSL        bool
	SkipTLSVerify bool
}

type LdapClient struct {
	LdapConfig LdapConfig
}

type ResolvedPermissions struct {
	IsAdmin    bool
	UIEnabled  bool
	APIEnabled bool
}

type LdapGroupMapping struct {
	ADGroup     string
	Permissions ResolvedPermissions
}
