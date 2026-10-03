package notification

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

// @Summary 		Get notification types
// @Description 	Get all supported notification types
// @Tags 			notification
// @Success 		200 {object} []NotificationTypeResponse
// @Router       	/notification/types [get]
// @Security 		BasicAuth
func (h *Handler) GetNotificationTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := h.Service.getNotificationTypes()
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: list})
	}
}

// @Summary 		Get notification config
// @Description 	Get notification config for current user
// @Tags 			notification
// @Success 		200 {object} NotificationConfigResponse
// @Router       	/user/notification [get]
// @Security 		BasicAuth
func (h *Handler) GetUserNotificationSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var n NotificationConfigResponse
		n, err := h.Service.getUserNotificationSettings(userID)
		if err != nil {
			h.Logger.Error("Failed to get user notification settings", "user", email, "domain", "notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get user notification"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: n})
	}
}

// @Summary 		Set notification config
// @Description 	Allows authenticated user to set notification config
// @Tags 			notification
// @Accept 			json
// @Param			request body NotificationConfigRequest true "Request body"
// @Success 		200 {object} bool
// @Failure      	400 {object} string
// @Router       	/user/notification [post]
// @Security 		BasicAuth
func (h *Handler) SetUserNotificationSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req NotificationConfigRequest
		json.NewDecoder(r.Body).Decode(&req)

		var resp shared.ApiResponse[any]
		if err := h.Service.setUserNotificationSettings(userID, req, ctx); err != nil {
			switch {
			case errors.Is(err, shared.ErrNotificationTypeNotSupported):
				h.Logger.Error("Notification type not supported", "user", email, "domain", "notification", "type", req.Type)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Notification type not supported",
				}

			case errors.Is(err, shared.ErrFieldMissing):
				h.Logger.Error("Required field missing", "user", email, "domain", "notification", "type", req.Type)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Required field missing",
				}

			case errors.Is(err, shared.ErrMissingAuthValues):
				h.Logger.Error("Missing username or password", "user", email, "domain", "notification", "type", req.Type)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing username or password",
				}

			default:
				h.Logger.Error("Failed to set user notification settings", "user", email, "domain", "notification", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to set user notification settings",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("User notification settings set", "user", email, "domain", "notification", "type", req.Type)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// @Summary 		Delete notification config
// @Description 	Allows authenticated user to delete notification config
// @Tags 			notification
// @Success 		200 {object} bool
// @Router       	/user/notification [delete]
// @Security 		BasicAuth
func (h *Handler) DeleteUserNotificationSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		if err := h.Service.deleteUserNotificationSettings(userID, ctx); err != nil {
			h.Logger.Error("Failed to delete user notification settings", "user", email, "domain", "notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to delete user notification"})
			return
		}

		h.Logger.Info("User notification settings deleted", "user", email, "domain", "notification")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// @Summary 		Test notification config
// @Description 	Allows authenticated user to test notification config and receive a message
// @Tags 			notification
// @Success 		200 {object} bool
// @Router       	/user/notification/test [post]
// @Security 		BasicAuth
func (h *Handler) QueueTestNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		if err := h.Service.queueTestNotification(userID); err != nil {
			h.Logger.Error("Failed to queue test notification", "user", email, "domain", "notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to queue test notification"})
			return
		}

		h.Logger.Info("Test notification queued", "user", email, "domain", "notification")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
