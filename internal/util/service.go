package util

import (
	"github.com/Masterminds/semver"
)

func (s *Service) getVersion() (VersionResponse, error) {
	defaultResp := VersionResponse{
		Version:   s.BuildInfo.Version,
		BuildDate: s.BuildInfo.BuildDate,
	}

	version, err := s.Repo.getVersion()
	if err != nil {
		return defaultResp, err
	}

	current, err := semver.NewVersion(s.BuildInfo.Version)
	if err != nil {
		return defaultResp, err
	}

	latest, err := semver.NewVersion(version.LatestVersion)
	if err != nil {
		return defaultResp, err
	}

	updateAvailable := false
	if latest.GreaterThan(current) {
		updateAvailable = true
	}

	resp := VersionResponse{
		Version:         s.BuildInfo.Version,
		BuildDate:       s.BuildInfo.BuildDate,
		UpdateAvailable: updateAvailable,
		LatestVersion:   version.LatestVersion,
		CheckedAt:       version.CheckedAt,
		ReleaseURL:      version.ReleaseURL,
	}

	return resp, nil
}
