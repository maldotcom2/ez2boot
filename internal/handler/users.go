package handler

import (
	"encoding/json"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"ez2boot/internal/service"
	"log/slog"
	"net/http"
)

// Handler to register new user
func RegisterUser(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u model.User
		err := json.NewDecoder(r.Body).Decode(&u)

		logger.Info("Attempted registration", "username", u.Username)

		if err != nil {
			logger.Info("Malformed request", "username", u.Username)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = service.ValidateAndCreateUser(repo, u); err != nil {
			logger.Info("Failed to create user", "username", u.Username, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
	}
}
