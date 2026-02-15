package util

func (r *Repository) getRelease() (LatestRelease, error) {
	var v LatestRelease
	if err := r.Base.DB.QueryRow("SELECT latest_release, latest_prerelease, checked_at, release_url, prerelease_url FROM release LIMIT 1").Scan(&v.LatestRelease, &v.LatestPreRelease, &v.CheckedAt, &v.ReleaseURL, &v.PreReleaseURL); err != nil {
		return LatestRelease{}, err
	}

	return v, nil
}

func (r *Repository) updateRelease(req RepoReleaseRequest) error {
	query := `INSERT INTO release (id, latest_release, latest_prerelease, checked_at, release_url, prerelease_url) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
			latest_release = EXCLUDED.latest_release,
			latest_prerelease = EXCLUDED.latest_prerelease,
			checked_at = EXCLUDED.checked_at,
			release_url = EXCLUDED.release_url,
			prerelease_url = EXCLUDED.prerelease_url`

	if _, err := r.Base.DB.Exec(query, 1, req.LatestRelease, req.LatestPreRelease, req.CheckedAt, req.ReleaseURL, req.PreReleaseURL); err != nil {
		return err
	}

	return nil
}
