package user

import (
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"log/slog"
)

type Repository struct {
	Base   *db.Repository
	Logger *slog.Logger
}

type Service struct {
	Repo   *Repository
	Config *model.Config
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
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserSession struct {
	SessionExpiry int64
	UserID        string
	Email         string
	Password      string
	Token         string
}
