package util

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base   *db.Repository
	Logger *slog.Logger
}

type Service struct {
	Repo      *Repository
	Config    *config.Config
	BuildInfo BuildInfo
	Logger    *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type BuildInfo struct {
	Version   string
	BuildDate string
}

// Used by Github polling job
type RepoReleaseRequest struct {
	LatestRelease    string `json:"latest_release"`
	LatestPreRelease string `json:"latest_prerelease"`
	CheckedAt        int64  `json:"checked_at"`
	ReleaseURL       string `json:"release_url"`
	PreReleaseURL    string `json:"prerelease_url"`
}

// Internal transport
type LatestRelease struct {
	LatestRelease    *string `json:"-"`
	LatestPreRelease *string `json:"-"`
	CheckedAt        *int64  `json:"-"`
	ReleaseURL       *string `json:"-"`
	PreReleaseURL    *string `json:"-"`
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	PreRelease bool   `json:"prerelease"`
}

// Used by UI
type VersionResponse struct {
	Version         string `json:"version"`
	BuildDate       string `json:"build_date"`
	UpdateAvailable bool   `json:"update_available"`
	LatestRelease   string `json:"latest_release"`
	CheckedAt       int64  `json:"checked_at"`
	ReleaseURL      string `json:"release_url"`
}
