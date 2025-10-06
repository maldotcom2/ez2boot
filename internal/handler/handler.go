package handler

import (
	"encoding/json"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"ez2boot/internal/utils"
	"log/slog"
	"net/http"
)

func GetServers(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := repo.GetServers(logger)
		if err != nil {
			logger.Error("Failed to get servers", "error", err)
			http.Error(w, "Failed to get servers", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(servers)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func GetSessions(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := repo.GetSessions(logger)
		if err != nil {
			logger.Error("Failed to get sessions", "error", err)
			http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(servers)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func NewSession(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group and email
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		// Validate TODO: MOVE THIS SOMEWHERE ELSE
		if session.Email == "" || session.ServerGroup == "" {
			logger.Error("Email and ServerGroup required")
			http.Error(w, "Email and ServerGroup required", http.StatusBadRequest)
		}

		// Generate token
		token, err := utils.GenerateToken(16)
		if err != nil {
			logger.Error("Failed to generate session token", "error", err)
			http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		}

		// Write session info to DB
		session.Token = token
		session, err = repo.NewSession(session, logger)
		if err != nil {
			logger.Error("Failed to create new session", "error", err)
			http.Error(w, "Failed to create new session", http.StatusInternalServerError)
			return
		}

		// Return to client
		session.Message = "Success"
		err = json.NewEncoder(w).Encode(session)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

/* func UpdateSession(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := repo.UpdateSession()
		if err != nil {
			logger.Error("Failed to update session", "error", err)
			http.Error(w, "Failed to update session", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(servers)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
} */
