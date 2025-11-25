package notification

import (
	"encoding/json"
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
