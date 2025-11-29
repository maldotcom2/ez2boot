package notification

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"net/http"
)

// Retrieves all supported notification types, used to list available options
func (h *Handler) GetNotificationTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := h.Service.getNotificationTypes()
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: list})
	}
}

func (h *Handler) GetUserNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var n NotificationRequest
		n, err := h.Service.getUserNotification(userID)
		if err != nil {
			h.Logger.Error("Failed to get user notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get user notification"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: n})
	}
}

func (h *Handler) SetUserNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var req NotificationRequest
		json.NewDecoder(r.Body).Decode(&req)

		var resp shared.ApiResponse[any]
		if err := h.Service.setUserNotification(userID, req); err != nil {
			switch {
			case errors.Is(err, shared.ErrNotificationTypeNotSupported):
				h.Logger.Error(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Notification type not supported",
				}

			case errors.Is(err, shared.ErrFieldMissing):
				h.Logger.Error("Required field missing", "error", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   err.Error(), // TODO is this ok?
				}

			default:
				h.Logger.Error("Failed to store notification preferences", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to store notification preferences",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("User notification config set")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
