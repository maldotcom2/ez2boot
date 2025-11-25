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
	Type() string                                    // Identifier
	Label() string                                   // UI label
	Send(msg string, title string, cfg string) error // Send the notification
}

type NotificationTypeRequest struct {
	Type  string `json:"type"`
	Label string `json:"label"`
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
