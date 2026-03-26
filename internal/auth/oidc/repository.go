package oidc

import "database/sql"

func (r *Repository) getOidcConfig() (OidcConfigStore, error) {
	var c OidcConfigStore

	query := `SELECT issuer_url, client_id, client_secret, app_url FROM oidc_config WHERE id = 1`

	if err := r.Base.DB.QueryRow(query).Scan(&c.IssuerURL, &c.ClientID, &c.ClientSecret, &c.AppURL); err != nil {
		return OidcConfigStore{}, err
	}

	return c, nil
}

func (r *Repository) setOidcConfig(req OidcConfigStore) error {
	query := `INSERT INTO Oidc_config (id, issuer_url, client_id, client_secret, app_url) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET issuer_url = EXCLUDED.issuer_url, client_id = EXCLUDED.client_id, 
			client_secret = EXCLUDED.client_secret, app_url = EXCLUDED.app_url`

	if _, err := r.Base.DB.Exec(query, 1, req.IssuerURL, req.ClientID, req.ClientSecret, req.AppURL); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getOidcSecret() ([]byte, error) {
	var encSecret []byte
	err := r.Base.DB.QueryRow("SELECT client_secret FROM oidc_config WHERE id = 1").Scan(&encSecret)
	return encSecret, err
}

func (r *Repository) setOidcSecretTx(tx *sql.Tx, encSecret []byte) error {
	_, err := tx.Exec("UPDATE oidc_config SET oidc_secret = $1 WHERE id = 1", encSecret)
	return err
}

func (r *Repository) deleteOidcConfig() error {
	if _, err := r.Base.DB.Exec("DELETE FROM oidc_config"); err != nil {
		return err
	}

	return nil
}
