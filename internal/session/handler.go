package session

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetServerSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := h.Service.getServerSessions()
		if err != nil {
			h.Logger.Error("Failed to get server sessions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get server sessions"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: sessions})
	}
}

func (h *Handler) GetServerSessionSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := h.Service.getServerSessionSummary()
		if err != nil {
			h.Logger.Error("Failed to get server session summary", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get server session summary"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: summary})
	}
}

func (h *Handler) NewServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var session ServerSession
		json.NewDecoder(r.Body).Decode(&session)
		session.UserID = userID

		// Create the session
		session, err := h.Service.newServerSession(session)
		if err != nil {
			h.Logger.Error("Failed to create new session", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to create new session"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: session})
	}
}

func (h *Handler) UpdateServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var session ServerSession
		json.NewDecoder(r.Body).Decode(&session)
		session.UserID = userID

		session, err := h.Service.updateServerSession(session) //TODO make this more clear
		if err != nil {
			if errors.Is(shared.ErrNoRowsUpdated, err) {
				h.Logger.Error("Requsted session for update was either not found or not owned", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to find session"})
				return
			}
			h.Logger.Error("Error while updating session", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while updating session"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: session})
	}
}
