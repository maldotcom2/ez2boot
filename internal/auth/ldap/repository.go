package ldap

import (
	"encoding/json"
)

func (r *Repository) getLdapConfig() (LdapConfigStore, error) {
	var c LdapConfigStore

	query := `SELECT host, port, base_dn, bind_dn, bind_password, use_ssl, skip_tls_verify FROM ldap_config WHERE id = 1`

	if err := r.Base.DB.QueryRow(query).Scan(&c.Host, &c.Port, &c.BaseDN, &c.BindDN, &c.EncBindPassword, &c.UseSSL, &c.SkipTLSVerify); err != nil {
		return LdapConfigStore{}, err
	}

	return c, nil
}

func (r *Repository) setLdapConfig(req LdapConfigStore) error {
	query := `INSERT INTO ldap_config (id, host, port, base_dn, bind_dn, bind_password, use_ssl, skip_tls_verify) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET host = EXCLUDED.host, port = EXCLUDED.port, base_dn = EXCLUDED.base_dn, bind_dn = EXCLUDED.bind_dn, 
			bind_password = EXCLUDED.bind_password, use_ssl = EXCLUDED.use_ssl, skip_tls_verify = EXCLUDED.skip_tls_verify`

	if _, err := r.Base.DB.Exec(query, 1, req.Host, req.Port, req.BaseDN, req.BindDN, req.EncBindPassword, req.UseSSL, req.SkipTLSVerify); err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleteLdapConfig() error {
	if _, err := r.Base.DB.Exec("DELETE from ldap_config"); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getGroupMappings() ([]LdapGroupMapping, error) {
	rows, err := r.Base.DB.Query("SELECT ad_group, permissions FROM ldap_group_mappings")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var mappings []LdapGroupMapping
	for rows.Next() {
		var m LdapGroupMapping
		var raw string
		if err := rows.Scan(&m.ADGroup, &raw); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(raw), &m.Permissions); err != nil {
			return nil, err
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}
