package oidc

import (
	"context"
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"log/slog"

	coreos "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type OidcProvider interface {
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	AuthCodeURL(state string) string
	VerifyIDToken(ctx context.Context, token *oauth2.Token) (map[string]any, error)
}

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo        *Repository
	UserService *user.Service
	Audit       *audit.Service
	Encryptor   Encryptor
	Provider    OidcProvider
	Logger      *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

// For read/write - contains encrypted secret
type OidcConfigStore struct {
	IssuerURL    string
	ClientID     string
	ClientSecret []byte
	RedirectURI  string
}

// For internal OIDC operations
type OidcConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Set Oidc config
type OidcConfigRequest struct {
	IssuerURL    string `json:"issuer_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

// Get current Oidc config for UI
type OidcConfigResponse struct {
	IssuerURL   string `json:"issuer_url"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
}

type OidcProviderImpl struct {
	provider  *coreos.Provider
	oauth2Cfg oauth2.Config
	verifier  *coreos.IDTokenVerifier
}
