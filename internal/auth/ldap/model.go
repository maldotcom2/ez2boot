package ldap

import (
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"log/slog"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo        *Repository
	UserService *user.Service
	Encryptor   Encryptor
	Logger      *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

// For read/write - contains encrypted password
type LdapConfigStore struct {
	Host            string
	Port            int64
	BaseDN          string
	BindDN          string
	EncBindPassword []byte
	UseSSL          bool
	SkipTLSVerify   bool
}

// For internal LDAP operations
type LdapConfig struct {
	Host          string
	Port          int64
	BaseDN        string
	BindDN        string
	BindPassword  string
	UseSSL        bool
	SkipTLSVerify bool
}

// Set Ldap config
type LdapConfigRequest struct {
	Host          string `json:"host"`
	Port          int64  `json:"port"`
	BaseDN        string `json:"base_dn"`
	BindDN        string `json:"bind_dn"`
	BindPassword  string `json:"bind_password"`
	UseSSL        bool   `json:"use_ssl"`
	SkipTLSVerify bool   `json:"skip_tls_verify"`
}

// Get current Ldap config for UI
type LdapConfigResponse struct {
	Host          string `json:"host"`
	Port          int64  `json:"port"`
	BaseDN        string `json:"base_dn"`
	BindDN        string `json:"bind_dn"`
	BindPassword  string `json:"bind_password"`
	UseSSL        bool   `json:"use_ssl"`
	SkipTLSVerify bool   `json:"skip_tls_verify"`
}

type LdapClient struct {
	LdapConfig LdapConfig
}

type LdapSearchRequest struct {
	Query string `json:"query"`
}

type LdapSearchResponse struct {
	DisplayName string
	Email       string
}

type CreateLdapUserRequest struct {
	Email string `json:"email"`
}
