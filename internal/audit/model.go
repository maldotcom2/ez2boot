package audit

import (
	"ez2boot/internal/db"
	"log/slog"
	"net/http"
)

type AdminChecker interface {
	UserIsAdmin(w http.ResponseWriter, r *http.Request) bool
}

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo   *Repository
	Logger *slog.Logger
}

type Handler struct {
	Service      *Service
	AdminChecker AdminChecker
	Logger       *slog.Logger
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
	Metadata     map[string]any
}

type AuditLogEvent struct {
	ActorUserID  int64
	ActorEmail   string
	TargetUserID int64
	TargetEmail  string
	Action       string
	Resource     string
	Success      bool
	Reason       string
	//Metadata     map[string]any
	TimeStamp int64
}

type AuditLogRequest struct {
	// Pagination
	Limit  int   `schema:"limit"`
	Before int64 `schema:"before"` // cursor (timestamp)

	// Filters
	ActorEmail  string `schema:"actor_email"`
	TargetEmail string `schema:"target_email"`
	Action      string `schema:"action"`
	Resource    string `schema:"resource"`
	Success     *bool  `schema:"success"`
	Reason      string `schema:"reason"`
	//Metadata    string `schema:"metadata"`

	// Time range
	From int64 `schema:"from"`
	To   int64 `schema:"to"`
}

type AuditLogResponse struct {
	Events     []AuditLogEvent `json:"events"`
	NextCursor *int64          `json:"next_cursor,omitempty"`
}
