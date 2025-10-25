package user

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"log/slog"
	"net/http"
)

// Handler to register new user
func RegisterUser(repo *db.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u model.User
		err := json.NewDecoder(r.Body).Decode(&u)

		logger.Info("Attempted registration", "username", u.Username)

		if err != nil {
			logger.Info("Malformed request", "username", u.Username)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = validateAndCreateUser(repo, u); err != nil {
			logger.Info("Failed to create user", "username", u.Username, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func ChangePassword(repo *db.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c model.ChangePasswordRequest
		err := json.NewDecoder(r.Body).Decode(&c)

		logger.Info("Attempted password change", "username", c.Username)

		if err != nil {
			logger.Error("Malformed request", "username", c.Username)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = changePasswordByUser(repo, c, logger); err != nil {
			logger.Error("Failed to change password for user", "username", c.Username, "error", err)

			if errors.Is(err, shared.ErrAuthenticationFailed) {
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				return
			} else if errors.Is(err, shared.ErrInvalidPassword) {
				http.Error(w, "Password did not match complexity requirements", http.StatusBadRequest)
				return
			} else {
				http.Error(w, "Failed to change password", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
