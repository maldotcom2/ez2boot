package util

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: h})
	}
}
