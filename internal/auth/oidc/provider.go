package oidc

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

func (p *OidcProviderImpl) AuthCodeURL(state string) string {
	return p.oauth2Cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *OidcProviderImpl) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth2Cfg.Exchange(ctx, code)
}

func (p *OidcProviderImpl) VerifyIDToken(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id_token: %w", err)
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	return claims, nil
}
