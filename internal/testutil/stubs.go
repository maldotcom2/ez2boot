package testutil

import "ez2boot/internal/auth/ldap"

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
