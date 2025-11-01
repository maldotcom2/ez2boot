package notification

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

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type Sender interface {
	Type() string                                    // Get the name
	Send(msg string, title string, cfg string) error // Send the notification
}

type Notification struct {
	Id    int64
	Msg   string
	Title string
	Type  string
	Cfg   string
}
