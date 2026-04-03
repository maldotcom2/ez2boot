package session

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"fmt"
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

		var req ServerSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "session", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		// Create the session
		var resp shared.ApiResponse[any]
		session, err := h.Service.newServerSession(req, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrFieldMissing):
				h.Logger.Error("Failed to create new session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing field in request",
				}
			case errors.Is(err, shared.ErrDurationTooLong):
				h.Logger.Error("Failed to create new session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Max session duration is %s", h.Config.MaxServerSessionDuration),
				}
			default:
				h.Logger.Error("Failed to create new session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to create new session",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Server session created", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: session})
	}
}

func (h *Handler) UpdateServerSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req ServerSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "session", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		var resp shared.ApiResponse[any]
		session, err := h.Service.updateServerSession(req, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Warn("Requested session for update was either not found or not owned", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to find session",
				}
			case errors.Is(err, shared.ErrFieldMissing):
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing field in request",
				}
			case errors.Is(err, shared.ErrDurationTooLong):
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Max session duration is %s", h.Config.MaxServerSessionDuration),
				}
			default:
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to update server session",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Server session updated", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: session})
	}
}

func (h *Handler) UpdateServerSessionAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req ServerSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "session", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		var resp shared.ApiResponse[any]
		session, err := h.Service.updateServerSessionAdmin(req, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Warn("Requested session for update was either not found or not owned", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to find session",
				}
			case errors.Is(err, shared.ErrFieldMissing):
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing field in request",
				}
			case errors.Is(err, shared.ErrDurationTooLong):
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Max session duration is %s", h.Config.MaxServerSessionDuration),
				}
			default:
				h.Logger.Error("Failed to update server session", "user", email, "domain", "session", "server_group", session.ServerGroup, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to update server session",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Server session updated", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: session})
	}
}
