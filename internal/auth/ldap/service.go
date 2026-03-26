package ldap

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"fmt"

	goldap "github.com/go-ldap/ldap/v3"
)

// Authenticate user from email and password or err on auth fail
func (s *Service) Authenticate(email string, password string) error {
	ldapCFG, err := s.getLdapConfigInternal()
	if err != nil {
		return err
	}

	conn, err := s.connect(ldapCFG)
	if err != nil {
		return fmt.Errorf("%w: %v", shared.ErrLDAPConnection, err)
	}
	defer conn.Close()

	// Authenticate user
	err = conn.Bind(email, password)
	if err != nil {
		return err
	}

	user, err := s.UserService.GetCredentialsByEmail(email)
	if err != nil {
		return err
	}

	// Update last login
	if err := s.UserService.UpdateLastLogin(user.UserID); err != nil {
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
			return LdapConfig{}, shared.ErrLDAPConfigNotFound
		}

		return LdapConfig{}, err
	}

	// Decrypt password
	passwordBytes, err := s.Encryptor.Decrypt([]byte(ldapCFG.EncBindPassword))
	if err != nil {
		return LdapConfig{}, err
	}

	return LdapConfig{
		Host:          ldapCFG.Host,
		Port:          ldapCFG.Port,
		BaseDN:        ldapCFG.BaseDN,
		BindDN:        ldapCFG.BindDN,
		BindPassword:  string(passwordBytes),
		UseSSL:        ldapCFG.UseSSL,
		SkipTLSVerify: ldapCFG.SkipTLSVerify,
	}, nil
}

func (s *Service) setLdapConfig(req LdapConfigRequest, ctx context.Context) (err error) {
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
			Resource:    "ldap config",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

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

// Return encrypted data for re-encryption
func (s *Service) GetLdapPassword() ([]byte, error) {
	return s.Repo.getLdapPassword()
}

// Write re-encrypted data
func (s *Service) SetLdapPasswordTx(tx *sql.Tx, encPassword []byte) error {
	return s.Repo.setLdapPasswordTx(tx, encPassword)
}

func (s *Service) deleteLdapConfig(ctx context.Context) (err error) {
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
			Resource:    "ldap config",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	return s.Repo.deleteLdapConfig()
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

func (s *Service) SearchUser(req LdapSearchRequest) (LdapSearchResponse, error) {
	ldapCFG, err := s.getLdapConfigInternal()
	if err != nil {
		return LdapSearchResponse{}, err
	}

	conn, err := s.connect(ldapCFG)
	if err != nil {
		return LdapSearchResponse{}, fmt.Errorf("%w: %v", shared.ErrLDAPConnection, err)
	}

	defer conn.Close()

	// Bind as service account to search
	if err = conn.Bind(ldapCFG.BindDN, ldapCFG.BindPassword); err != nil {
		return LdapSearchResponse{}, err
	}

	searchRequest := goldap.NewSearchRequest(
		ldapCFG.BaseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(mail=%s*)", // Target user requires a mail field
			goldap.EscapeFilter(req.Query),
		),
		[]string{"cn", "mail"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return LdapSearchResponse{}, err
	}

	if len(result.Entries) == 0 {
		return LdapSearchResponse{}, shared.ErrUserNotFound
	}

	// Singular return - possibly expand to a collection
	entry := result.Entries[0]
	return LdapSearchResponse{
		DisplayName: entry.GetAttributeValue("cn"),
		Email:       entry.GetAttributeValue("mail"),
	}, nil
}

func (s *Service) createLdapUser(email string, ctx context.Context) error {
	req := LdapSearchRequest{
		Query: email,
	}

	// Check user exists - no user returns an err
	if _, err := s.Searcher.SearchUser(req); err != nil {
		return err
	}

	// Create user
	if _, err := s.UserService.CreateExternalUser(email, shared.IdentityProviderLDAP, ctx); err != nil {
		return err
	}

	return nil
}
