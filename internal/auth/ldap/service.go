package ldap

import (
	"crypto/tls"
	"fmt"

	goldap "github.com/go-ldap/ldap/v3"
)

func (s *Service) Authenticate(upn string, password string) ([]string, error) {
	ldapCFG, err := s.Repo.getLdapConfig()
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(ldapCFG)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Authenticate user
	err = conn.Bind(upn, password)
	if err != nil {
		return nil, err
	}

	// Re-bind as service account to search
	err = conn.Bind(ldapCFG.BindDN, ldapCFG.BindPassword)
	if err != nil {
		return nil, err
	}

	// Search group memberships
	groups, err := s.getADGroupMembership(conn, ldapCFG.BaseDN, upn)
	if err != nil {
		return nil, err
	}

	return groups, nil
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

func (s *Service) getADGroupMembership(conn *goldap.Conn, baseDN string, upn string) ([]string, error) {
	searchRequest := goldap.NewSearchRequest(
		baseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(userPrincipalName=%s)", goldap.EscapeFilter(upn)),
		[]string{"memberOf"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found in directory: %s", upn)
	}

	return result.Entries[0].GetAttributeValues("memberOf"), nil
}

func (s *Service) ResolvePermissions(groups []string) (ResolvedPermissions, error) {
	// Get configured mapping of AD group names to roles
	mappings, err := s.Repo.getGroupMappings()
	if err != nil {
		return ResolvedPermissions{}, err
	}

	// Iterate mapping and identify which roles user should have
	var resolved ResolvedPermissions
	for _, group := range groups {
		for _, mapping := range mappings {
			if mapping.ADGroup == group {
				if mapping.Permissions.IsAdmin {
					resolved.IsAdmin = true
				}
				if mapping.Permissions.UIEnabled {
					resolved.UIEnabled = true
				}
				if mapping.Permissions.APIEnabled {
					resolved.APIEnabled = true
				}
			}
		}
	}

	return resolved, nil
}
