package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
)

// Get version info for UI
func (s *Service) getVersion() (VersionResponse, error) {
	resp := VersionResponse{
		Version:   s.BuildInfo.Version,
		BuildDate: s.BuildInfo.BuildDate,
	}

	release, err := s.Repo.getRelease()
	if err != nil {
		return resp, err
	}

	currentVersion, err := semver.NewVersion(s.BuildInfo.Version)
	if err != nil {
		return resp, err
	}

	if release.CheckedAt != nil {
		resp.CheckedAt = *release.CheckedAt
	}

	var latestRelease, latestPreRelease *semver.Version

	// Parse latest release
	if release.LatestRelease != nil && *release.LatestRelease != "" {
		latestRelease, err = semver.NewVersion(*release.LatestRelease)
		if err != nil {
			return resp, err
		}
	}

	// Parse latest prerelease
	if release.LatestPreRelease != nil && *release.LatestPreRelease != "" {
		latestPreRelease, err = semver.NewVersion(*release.LatestPreRelease)
		if err != nil {
			return resp, err
		}
	}

	// Select candidate version
	var candidate *semver.Version

	if latestRelease != nil {
		candidate = latestRelease
	}

	if s.Config.ShowBetaVersions && latestPreRelease != nil && (candidate == nil || latestPreRelease.GreaterThan(candidate)) {
		candidate = latestPreRelease
	}

	// Determine candidate tag and URL - defensive against nil pointer dereference
	if candidate != nil {
		switch candidate {
		case latestRelease:
			if release.LatestRelease != nil {
				resp.LatestRelease = *release.LatestRelease
			}
			if release.ReleaseURL != nil {
				resp.ReleaseURL = *release.ReleaseURL
			}
		case latestPreRelease:
			if release.LatestPreRelease != nil {
				resp.LatestRelease = *release.LatestPreRelease
			}
			if release.PreReleaseURL != nil {
				resp.ReleaseURL = *release.PreReleaseURL
			}
		}

		if candidate.GreaterThan(currentVersion) {
			resp.UpdateAvailable = true
		}
	}

	return resp, nil
}

// Check repo for latest version and update DB
func (s *Service) CheckRelease() error {
	url := "https://api.github.com/repos/maldotcom2/ez2boot/releases"

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
	var ghReleases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghReleases); err != nil {
		return err
	}

	var latestRelease, latestPreRelease GitHubRelease
	var foundRelease, foundPreRelease bool

	// Assumes releases retrieved chronologically
	for _, r := range ghReleases {
		if r.PreRelease {
			if !foundPreRelease {
				latestPreRelease = r
				foundPreRelease = true
			}
		} else {
			if !foundRelease {
				latestRelease = r
				foundRelease = true
			}
		}

		if foundRelease && foundPreRelease {
			break //found
		}
	}

	req := RepoReleaseRequest{
		CheckedAt: time.Now().Unix(),
	}

	if foundRelease {
		req.LatestRelease = latestRelease.TagName
		req.ReleaseURL = latestRelease.HTMLURL
	}

	if foundPreRelease {
		req.LatestPreRelease = latestPreRelease.TagName
		req.PreReleaseURL = latestPreRelease.HTMLURL
	}

	// Write to DB
	if err := s.Repo.updateRelease(req); err != nil {
		return err
	}

	return nil
}
