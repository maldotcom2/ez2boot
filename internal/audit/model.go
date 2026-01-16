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
	UserID   int64
	Email    string
	Action   string
	Resource string
	Result   string
}
