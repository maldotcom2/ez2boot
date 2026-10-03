package audit

import (
	"encoding/json"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"

	"github.com/gorilla/schema"
)

// @Summary 		Get audit events
// @Description 	[Admin] retrieve paginated list of audit events, filtered by actors, actions or time
// @Tags 			audit
// @Param        	limit         query    int    false  "Number of records to return" default(50)
// @Param        	before        query    int64  false  "Cursor pagination token represented as a Unix timestamp"
// @Param        	actor_email   query    string false  "Filter logs by the email of the actor" format(email)
// @Param        	target_email  query    string false  "Filter logs by the email of the target user" format(email)
// @Param        	action        query    string false  "Filter by action type (e.g., login, server_start)"
// @Param        	resource      query    string false  "Filter by impacted cloud resource ID"
// @Param        	success       query    bool   false  "Filter by event execution status (true/false)"
// @Param        	reason        query    string false  "Filter by failure reasons or details"
// @Param        	metadata      query    string false  "Search text within metadata payloads"
// @Param        	from          query    int64  false  "Fetch events after this Unix timestamp"
// @Param        	to            query    int64  false  "Fetch events before this Unix timestamp"
// @Success 		200 {object} AuditLogResponse
// @Failure      	400 {object} string
// @Router       	/audit/events [get]
// @Security 		BasicAuth
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
