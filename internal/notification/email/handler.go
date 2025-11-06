package email

import (
	"encoding/json"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) AddOrUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var c Config
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			h.Logger.Error("Malformed request")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		if err := h.Service.AddOrUpdate(userID, c); err != nil {
			h.Logger.Error("Failed to add or update email notifications", "userID", userID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to create new session"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
