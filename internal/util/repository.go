package util

func (r *Repository) getVersion() (VersionResponse, error) {
	var v VersionResponse
	if err := r.Base.DB.QueryRow("SELECT latest_version, checked_at, release_url FROM version LIMIT 1").Scan(&v.LatestVersion, &v.CheckedAt, &v.ReleaseURL); err != nil {
		return VersionResponse{}, err
	}

	return v, nil
}

func (r *Repository) updateVersion(req RepoVersionRequest) error {
	query := `INSERT INTO version (id, latest_version, checked_at, release_url) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
			latest_version = EXCLUDED.latest_version,
			checked_at = EXCLUDED.checked_at,
			release_url = EXCLUDED.release_url`

	if _, err := r.Base.DB.Exec(query, 1, req.LatestVersion, req.CheckedAt, req.ReleaseURL); err != nil {
		return err
	}

	return nil
}
