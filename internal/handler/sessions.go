package handler

import (
	"encoding/json"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"net/http"
)

func GetSessions(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := repo.GetSessions()
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
		// Receive server_group, email and duration
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		// Validate TODO: MOVE THIS SOMEWHERE ELSE
		if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
			logger.Error("Email server_group and duration required")
			http.Error(w, "Email and ServerGroup required", http.StatusBadRequest)
		}

		// Generate token
		token, err := GenerateToken(16)
		if err != nil {
			logger.Error("Failed to generate session token", "error", err)
			http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		}

		// Write session info to DB
		session.Token = token
		session, err = repo.NewSession(session)
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

func UpdateSession(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group, email and duration
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		// Validate TODO: MOVE THIS SOMEWHERE ELSE
		if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
			logger.Error("Email, server_group and duration required")
			http.Error(w, "Email, server_group and duration required", http.StatusBadRequest)
		}

		updated, session, err := repo.UpdateSession(session)
		if err != nil {
			logger.Error("Failed to update session", "error", err)
			http.Error(w, "Failed to update session", http.StatusInternalServerError)
			return
		}

		if !updated {
			logger.Error("Session not found", "error", err)
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(session)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
