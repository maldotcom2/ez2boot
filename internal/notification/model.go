package notification

import (
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo     *Repository
	Logger   *slog.Logger
	Handlers map[string]ConfigHandler // type specific handlers
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

// Notification channels must implement this
type ConfigHandler interface {
	Validate(cfg map[string]any) error           // Type specific validation
	ToConfig(cfg map[string]any) (string, error) // Marshal into json
}

// Used by UI to populate options
type NotificationTypeRequest struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// User stored notification preferences
type NotificationUpdateRequest struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
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
