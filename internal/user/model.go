package user

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base   *db.Repository
	Logger *slog.Logger
}

type Service struct {
	Repo   *Repository
	Config *config.Config
	Audit  *audit.Service
	Logger *slog.Logger
}

type Handler struct {
	Service *Service
	Config  *config.Config
	Logger  *slog.Logger
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserSession struct {
	UserID        int64
	SessionExpiry int64
	Email         string
	Token         string
}

type UserSessionResponse struct {
	UserID        int64
	SessionExpiry int64
	Email         string
}

// For get request only
type User struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email"`
	IsActive   bool   `json:"is_active"`
	IsAdmin    bool   `json:"is_admin"`
	APIEnabled bool   `json:"api_enabled"`
	UIEnabled  bool   `json:"ui_enabled"`
	LastLogin  *int64 `json:"last_login"`
}

// Used for admin panel user updates
type UpdateUserRequest struct {
	UserID     int64 `json:"user_id"`
	IsActive   bool  `json:"is_active"`
	IsAdmin    bool  `json:"is_admin"`
	APIEnabled bool  `json:"api_enabled"`
	UIEnabled  bool  `json:"ui_enabled"`
}

type DeleteUserRequest struct {
	UserID int64 `json:"user_id"`
}

type CreateUserRequest struct {
	UserID     int64  `json:"-"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsActive   bool   `json:"is_active"`
	IsAdmin    bool   `json:"is_admin"`
	APIEnabled bool   `json:"api_enabled"`
	UIEnabled  bool   `json:"ui_enabled"`
}

// Intermediate stuct used after password hashing
type CreateUser struct {
	UserID       int64
	Email        string
	PasswordHash string
	IsActive     bool
	IsAdmin      bool
	APIEnabled   bool
	UIEnabled    bool
}

type UserAuthResponse struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email"`
	IsActive   bool   `json:"is_active"`
	IsAdmin    bool   `json:"is_admin"`
	APIEnabled bool   `json:"api_enabled"`
	UIEnabled  bool   `json:"ui_enabled"`
}

type SetupResponse struct {
	SetupMode bool `json:"setup_mode"`
}
