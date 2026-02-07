package util

func (r *Repository) getVersion() (VersionResponse, error) {
	var v VersionResponse
	if err := r.Base.DB.QueryRow("SELECT latest_version, checked_at, release_url FROM version LIMIT 1").Scan(&v.LatestVersion, &v.CheckedAt, &v.ReleaseURL); err != nil {
		return VersionResponse{}, err
	}

	return v, nil
}
