package util

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, err := h.Service.getVersion()
		if err != nil {
			h.Logger.Warn("Error while getting latest version", "error", err)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Data: version, Error: "Error while getting latest version"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: version})
	}
}
