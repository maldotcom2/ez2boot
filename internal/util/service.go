package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver"
)

// Get version info for UI
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

// Check repo for latest version and update DB
func (s *Service) UpdateVersion() error {
	url := "https://api.github.com/repos/maldotcom2/ez2boot/releases/latest"

	ghReq, _ := http.NewRequest("GET", url, nil)
	ghReq.Header.Set("User-Agent", "ez2boot")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(ghReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub returned %s for url %s", resp.Status, url)
	}

	// Decode relevant fields from response
	var ghRelease GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return err
	}

	// Create targetted struct
	req := RepoVersionRequest{
		LatestVersion: ghRelease.TagName,
		CheckedAt:     time.Now().Unix(),
		ReleaseURL:    ghRelease.HTMLURL,
	}

	// Write it
	if err := s.Repo.updateVersion(req); err != nil {
		return err
	}

	return nil
}
