package audit

import (
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo   *Repository
	Logger *slog.Logger
}

type Event struct {
	ActorUserID  int64
	ActorEmail   string
	TargetUserID int64
	TargetEmail  string
	Action       string
	Resource     string
	Success      bool
	Reason       string
}
