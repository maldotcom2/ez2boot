package oidc

import (
	"context"
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"fmt"
	"log/slog"

	coreos "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func NewHandler(oidcService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: oidcService,
		Logger:  logger,
	}
}

func NewService(oidcRepo *Repository, userService *user.Service, audit *audit.Service, encryptor Encryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:        oidcRepo,
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

func NewOidcProvider(ctx context.Context, cfg OidcConfig) (OidcProvider, error) {
	provider, err := coreos.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise OIDC provider: %w", err)
	}

	oauth2Cfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{coreos.ScopeOpenID, "email", "profile"},
	}

	verifier := provider.Verifier(&coreos.Config{
		ClientID: cfg.ClientID,
	})

	return &OidcProviderImpl{
		provider:  provider,
		oauth2Cfg: oauth2Cfg,
		verifier:  verifier,
	}, nil
}
