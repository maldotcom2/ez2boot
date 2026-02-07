package util

import (
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base   *db.Repository
	Logger *slog.Logger
}

type Service struct {
	Repo      *Repository
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

type VersionResponse struct {
	Version         string `json:"version"`
	BuildDate       string `json:"build_date"`
	UpdateAvailable bool   `json:"update_available"`
	LatestVersion   string `json:"latest_version"`
	CheckedAt       int64  `json:"checked_at"`
	ReleaseURL      string `json:"release_url"`
}
