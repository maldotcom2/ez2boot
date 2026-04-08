package testutil

import (
	"context"
	"ez2boot/internal/auth/ldap"

	"golang.org/x/oauth2"
)

// LDAP
type StubLdapService struct {
	AuthenticateFunc func(email, password string) error
}

type StubLdapSearcher struct {
	SearchUserFunc func(req ldap.LdapSearchRequest) (ldap.LdapSearchResponse, error)
}

func (m *StubLdapService) Authenticate(email, password string) error {
	return m.AuthenticateFunc(email, password)
}

func (s *StubLdapSearcher) SearchUser(req ldap.LdapSearchRequest) (ldap.LdapSearchResponse, error) {
	return s.SearchUserFunc(req)
}

// OIDC
type StubOidcProvider struct {
	ExchangeFunc      func(ctx context.Context, code string) (*oauth2.Token, error)
	AuthCodeURLFunc   func(state string) string
	VerifyIDTokenFunc func(ctx context.Context, token *oauth2.Token) (map[string]any, error)
}

func (s *StubOidcProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.ExchangeFunc(ctx, code)
}

func (s *StubOidcProvider) AuthCodeURL(state string) string {
	return s.AuthCodeURLFunc(state)
}

func (s *StubOidcProvider) VerifyIDToken(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
	return s.VerifyIDTokenFunc(ctx, token)
}

// Notifications
type StubNotificationChannel struct {
    SendFunc func(msg string, title string, cfg string) error
    Calls    []string // track what was sent
}

func (s *StubNotificationChannel) Type() string  { return "stub" }
func (s *StubNotificationChannel) Label() string { return "Stub" }
func (s *StubNotificationChannel) Send(msg string, title string, cfg string) error {
    s.Calls = append(s.Calls, title)
    if s.SendFunc != nil {
        return s.SendFunc(msg, title, cfg)
    }
    return nil
}
func (s *StubNotificationChannel) Validate(cfg map[string]any) error { return nil }
func (s *StubNotificationChannel) ToConfig(cfg map[string]any) (string, error) { return "{}", nil }
