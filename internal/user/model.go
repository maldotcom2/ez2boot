package user

import (
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
	Logger *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	UserID      int64
	Email       string
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserSession struct {
	UserID        int64
	SessionExpiry int64
	Email         string
	Token         string
}

type CreateUserRequest struct {
	UserID     int64
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsActive   bool   `json:"is_active"`
	IsAdmin    bool   `json:"is_admin"`
	APIEnabled bool   `json:"api_enabled"`
	UIEnabled  bool   `json:"ui_enabled"`
}

type CreateUser struct {
	UserID       int64
	Email        string
	PasswordHash string
	IsActive     bool
	IsAdmin      bool
	APIEnabled   bool
	UIEnabled    bool
}

type User struct {
	UserID     int64
	Email      string
	IsActive   bool
	IsAdmin    bool
	APIEnabled bool
	UIEnabled  bool
}

type SetupResponse struct {
	SetupMode bool
}
