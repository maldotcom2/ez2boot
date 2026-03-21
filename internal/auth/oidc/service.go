package oidc

import (
	"context"
	"database/sql"
	"errors"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
)

// UI calls, nulls password value
func (s *Service) getOidcConfig() (OidcConfigResponse, error) {
	oidcCFG, err := s.Repo.getOidcConfig()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OidcConfigResponse{}, shared.ErrOIDCConfigNotFound
		}

		return OidcConfigResponse{}, err
	}

	return OidcConfigResponse{
		IssuerURL:    oidcCFG.IssuerURL,
		ClientID:     oidcCFG.ClientID,
		ClientSecret: "",
		RedirectURI:  oidcCFG.RedirectURI,
	}, nil
}

// System calls, preserves password value
func (s *Service) getOidcConfigInternal() (OidcConfig, error) {
	oidcCFG, err := s.Repo.getOidcConfig()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OidcConfig{}, nil
		}

		return OidcConfig{}, err
	}

	// Decrypt secret
	secretBytes, err := s.Encryptor.Decrypt([]byte(oidcCFG.ClientSecret))
	if err != nil {
		return OidcConfig{}, err
	}

	return OidcConfig{
		IssuerURL:    oidcCFG.IssuerURL,
		ClientID:     oidcCFG.ClientID,
		ClientSecret: string(secretBytes),
		RedirectURI:  oidcCFG.RedirectURI,
	}, nil
}

func (s *Service) setOidcConfig(req OidcConfigRequest, ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "set",
			Resource:    "oidc config",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	// Encrypt secret
	encryptedBytes, err := s.Encryptor.Encrypt([]byte(req.ClientSecret))
	if err != nil {
		return err
	}

	c := OidcConfigStore{
		IssuerURL:    req.IssuerURL,
		ClientID:     req.ClientID,
		ClientSecret: encryptedBytes,
		RedirectURI:  req.RedirectURI,
	}

	if err = s.Repo.setOidcConfig(c); err != nil {
		return err
	}

	return nil
}

// Return encrypted data for re-encryption
func (s *Service) GetOidcSecret() ([]byte, error) {
	return s.Repo.getOidcSecret()
}

// Write re-encrypted data
func (s *Service) SetOidcSecretTx(tx *sql.Tx, encSecret []byte) error {
	return s.Repo.setOidcSecretTx(tx, encSecret)
}

func (s *Service) deleteOidcConfig(ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "delete",
			Resource:    "oidc config",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	return s.Repo.deleteOidcConfig()
}
