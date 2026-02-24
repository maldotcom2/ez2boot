package session

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetServerSessionSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		summary, err := h.Service.getServerSessionSummary()
		if err != nil {
			h.Logger.Error("Failed to get server session summary", "user", email, "domain", "session", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get server session summary"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: summary})
	}
}

func (h *Handler) NewServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var session ServerSessionRequest
		json.NewDecoder(r.Body).Decode(&session)
		session.UserID = userID

		// Create the session
		expiry, err := h.Service.newServerSession(session, ctx)
		if err != nil {
			h.Logger.Error("Failed to create new session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to create new session"})
			return
		}

		res := ServerSessionResponse{
			ServerGroup: session.ServerGroup,
			Duration:    session.Duration,
			Expiry:      expiry,
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: res})
	}
}

func (h *Handler) UpdateServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var session ServerSessionRequest
		json.NewDecoder(r.Body).Decode(&session)
		session.UserID = userID

		expiry, err := h.Service.updateServerSession(session, ctx)
		if err != nil {
			if errors.Is(err, shared.ErrNoRowsUpdated) {
				h.Logger.Warn("Requsted session for update was either not found or not owned", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to find session"})
				return
			}
			h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to update server session"})
			return
		}

		res := ServerSessionResponse{
			ServerGroup: session.ServerGroup,
			Duration:    session.Duration,
			Expiry:      expiry,
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: res})
	}
}
