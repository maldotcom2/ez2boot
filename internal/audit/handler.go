package audit

import (
	"encoding/json"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"

	"github.com/gorilla/schema"
)

func (h *Handler) GetAuditEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req AuditLogRequest

		decoder := schema.NewDecoder()
		decoder.IgnoreUnknownKeys(true)

		// Parse query values into struct
		if err := decoder.Decode(&req, r.URL.Query()); err != nil {
			h.Logger.Error("Failed to decode request", "user", email, "domain", "audit", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{
				Success: false,
				Error:   "Invalid query parameters",
			})
			return
		}

		// Admin check
		if !h.AdminChecker.UserIsAdmin(w, r) {
			return // response written by helper
		}

		events, err := h.Service.GetAuditEvents(req)
		if err != nil {
			h.Logger.Error("Failed to fetch audit events", "user", email, "domain", "audit", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{
				Success: false,
				Error:   "Failed to fetch audit events",
			})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[AuditLogResponse]{
			Success: true,
			Data:    events,
		})
	}
}
