package email

import (
	"encoding/json"
	"ez2boot/internal/contextkey"
	"net/http"
)

func (h *Handler) AddOrUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		var c Config
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			h.Logger.Error("Invalid request")
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := h.Service.AddOrUpdate(userID, c); err != nil {
			h.Logger.Error("Failed to add or update email notifications", "userID", userID, "error", err)
			http.Error(w, "Failed to create new session", http.StatusInternalServerError)
			return
		}
	}
}
