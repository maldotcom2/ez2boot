package session

import (
	"encoding/json"
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"log/slog"
	"net/http"
)

func GetSessions(repo *db.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := GetAllSessions()
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

func NewSession(repo *db.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group, email and duration
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		// Create the session
		session, err := createNewSession(session)
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

func UpdateSession(repo *db.Repository, logger *slog.Logger) http.HandlerFunc {
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
