package ldap

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"ez2boot/internal/shared"
	"fmt"

	goldap "github.com/go-ldap/ldap/v3"
)

// Authenticate user from upn and password, returning AD group membership or err on auth fail
func (s *Service) Authenticate(upn string, password string) error {
	ldapCFG, err := s.getLdapConfigInternal()
	if err != nil {
		return err
	}

	conn, err := s.connect(ldapCFG)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Authenticate user
	err = conn.Bind(upn, password)
	if err != nil {
		return err
	}

	// Re-bind as service account to search
	err = conn.Bind(ldapCFG.BindDN, ldapCFG.BindPassword)
	if err != nil {
		return err
	}

	return nil
}

// UI calls, nulls password value
func (s *Service) getLdapConfig() (LdapConfigResponse, error) {
	ldapCFG, err := s.Repo.getLdapConfig()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LdapConfigResponse{}, shared.ErrLDAPConfigNotFound
		}

		return LdapConfigResponse{}, err
	}

	return LdapConfigResponse{
		Host:          ldapCFG.Host,
		Port:          ldapCFG.Port,
		BaseDN:        ldapCFG.BaseDN,
		BindDN:        ldapCFG.BindDN,
		BindPassword:  "",
		UseSSL:        ldapCFG.UseSSL,
		SkipTLSVerify: ldapCFG.SkipTLSVerify,
	}, nil
}

// System calls, preserves password value
func (s *Service) getLdapConfigInternal() (LdapConfig, error) {
	ldapCFG, err := s.Repo.getLdapConfig()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LdapConfig{}, nil
		}

		return LdapConfig{}, err
	}

	// Decrypt password
	password, err := s.Encryptor.Decrypt([]byte(ldapCFG.EncBindPassword))
	if err != nil {
		return LdapConfig{}, err
	}

	return LdapConfig{
		Host:          ldapCFG.Host,
		Port:          ldapCFG.Port,
		BaseDN:        ldapCFG.BaseDN,
		BindDN:        ldapCFG.BindDN,
		BindPassword:  string(password),
		UseSSL:        ldapCFG.UseSSL,
		SkipTLSVerify: ldapCFG.SkipTLSVerify,
	}, nil
}

func (s *Service) setLdapConfig(req LdapConfigRequest) error {
	// Encrypt password
	encryptedBytes, err := s.Encryptor.Encrypt([]byte(req.BindPassword))
	if err != nil {
		return err
	}

	c := LdapConfigStore{
		Host:            req.Host,
		Port:            req.Port,
		BaseDN:          req.BaseDN,
		BindDN:          req.BindDN,
		EncBindPassword: encryptedBytes,
		UseSSL:          req.UseSSL,
		SkipTLSVerify:   req.SkipTLSVerify,
	}

	if err = s.Repo.setLdapConfig(c); err != nil {
		return err
	}

	return nil
}

func (s *Service) deleteLdapConfig() error {
	if err := s.Repo.deleteLdapConfig(); err != nil {
		return err
	}

	return nil
}

// Opens a connection to the LDAP server
func (s *Service) connect(ldapCFG LdapConfig) (*goldap.Conn, error) {
	if ldapCFG.UseSSL {
		addr := fmt.Sprintf("ldaps://%s:%d", ldapCFG.Host, ldapCFG.Port)
		return goldap.DialURL(addr, goldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: ldapCFG.SkipTLSVerify,
		}),
		)
	}

	addr := fmt.Sprintf("ldap://%s:%d", ldapCFG.Host, ldapCFG.Port)
	return goldap.DialURL(addr)
}

func (s *Service) SearchUser(req LdapSearchRequest) (LdapUser, error) {
	ldapCFG, err := s.getLdapConfigInternal()
	if err != nil {
		return LdapUser{}, err
	}

	conn, err := s.connect(ldapCFG)
	if err != nil {
		return LdapUser{}, err
	}

	defer conn.Close()

	// Bind as service account to search
	if err = conn.Bind(ldapCFG.BindDN, ldapCFG.BindPassword); err != nil {
		return LdapUser{}, err
	}

	searchRequest := goldap.NewSearchRequest(
		ldapCFG.BaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(mail=%s*)",
			goldap.EscapeFilter(req.Query),
		),
		[]string{"cn", "mail"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return LdapUser{}, err
	}

	if len(result.Entries) == 0 {
		return LdapUser{}, shared.ErrUserNotFound
	}

	// Test a singular return - possibly expand to a collection
	entry := result.Entries[0]
	return LdapUser{
		DisplayName: entry.GetAttributeValue("cn"),
		Email:       entry.GetAttributeValue("mail"),
	}, nil
}
