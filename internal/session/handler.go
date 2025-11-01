package session

import (
	"encoding/json"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetServerSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := h.Service.getServerSessions()
		if err != nil {
			h.Logger.Error("Failed to get sessions", "error", err)
			http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) NewServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		var s ServerSession
		json.NewDecoder(r.Body).Decode(&s)
		s.UserID = userID

		// Create the session
		s, err := h.Service.newServerSession(s)
		if err != nil {
			h.Logger.Error("Failed to create new session", "error", err)
			http.Error(w, "Failed to create new session", http.StatusInternalServerError)
			return
		}

		// Return to client
		s.Message = "Success"
		err = json.NewEncoder(w).Encode(s)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) UpdateServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Receive server_group, email and duration
		var s ServerSession
		json.NewDecoder(r.Body).Decode(&s)

		s, err := h.Service.updateServerSession(s)
		if err != nil {
			if err == shared.ErrSessionNotFound {
				h.Logger.Error("Failed to find session", "error", err)
				http.Error(w, "Failed to find session", http.StatusUnauthorized)
				return
			}
		}

		err = json.NewEncoder(w).Encode(s)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
