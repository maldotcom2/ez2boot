package notification

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"log/slog"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo      *Repository
	Audit     *audit.Service
	Encryptor Encryptor
	Logger    *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type NotificationChannel interface {
	Type() string                                    // Identifier
	Label() string                                   // UI label
	Send(msg string, title string, cfg string) error // Send the notification
	Validate(cfg map[string]any) error               // Type specific validation
	ToConfig(cfg map[string]any) (string, error)     // Marshal into json
}

// Used by UI to populate options
type NotificationTypeResponse struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// Internal transport only
type NotificationSetting struct {
	UserID    int64
	Type      string
	EncConfig []byte
}

// Used to save user notification config
type NotificationConfigRequest struct {
	Type          string         `json:"type"`
	ChannelConfig map[string]any `json:"channel_config"`
}

// Used to return currently populated user notification config
type NotificationConfigResponse struct {
	Type          string         `json:"type"`
	ChannelConfig map[string]any `json:"channel_config"`
}

type NewNotification struct {
	UserID int64
	Msg    string
	Title  string
	Time   int64
}

// Transport for notification processing
type Notification struct {
	UserID    int64
	Id        int64
	Msg       string
	Title     string
	Type      string
	Time      int64
	EncConfig []byte
}

type RotateEncryptionPhraseRequest struct {
	Phrase string `json:"phrase"`
}
