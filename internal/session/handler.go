package session

import (
	"encoding/json"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := h.Service.GetSessions()
		if err != nil {
			h.Logger.Error("Failed to get sessions", "error", err)
			http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) NewSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group, email and duration
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		// Create the session
		session, err := h.Service.createNewSession(session)
		if err != nil {
			h.Logger.Error("Failed to create new session", "error", err)
			http.Error(w, "Failed to create new session", http.StatusInternalServerError)
			return
		}

		// Return to client
		session.Message = "Success"
		err = json.NewEncoder(w).Encode(session)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) UpdateSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group, email and duration
		var session model.Session
		json.NewDecoder(r.Body).Decode(&session)

		session, err := h.Service.UpdateSession(session)
		if err != nil {
			if err == shared.ErrSessionNotFound {
				h.Logger.Error("Failed to find session", "error", err)
				http.Error(w, "Failed to find session", http.StatusUnauthorized)
			}
		}

		err = json.NewEncoder(w).Encode(session)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
