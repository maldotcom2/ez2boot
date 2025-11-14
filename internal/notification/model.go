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

// Notification channels must implement this
type Sender interface {
	Type() string                                    // Get the name
	Send(msg string, title string, cfg string) error // Send the notification
}

type NewNotification struct {
	UserID int64
	Msg    string
	Title  string
	Time   int64
}

type Notification struct {
	Id    int64
	Msg   string
	Title string
	Type  string
	Time  int64
	Cfg   string
}
