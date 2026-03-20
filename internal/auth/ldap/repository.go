package ldap

import "database/sql"

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

func (r *Repository) getLdapPassword() ([]byte, error) {
	var encPassword []byte
	err := r.Base.DB.QueryRow(`SELECT bind_password FROM ldap_config WHERE id = 1`).Scan(&encPassword)
	return encPassword, err
}

func (r *Repository) setLdapPasswordTx(tx *sql.Tx, encPassword []byte) error {
	_, err := tx.Exec("UPDATE ldap_config SET bind_password = $1 WHERE id = 1", encPassword)
	return err
}

func (r *Repository) deleteLdapConfig() error {
	if _, err := r.Base.DB.Exec("DELETE FROM ldap_config"); err != nil {
		return err
	}

	return nil
}
