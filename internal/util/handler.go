package util

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"net/http"
)

// @Summary 		Get version
// @Description 	Allow authenticated user to get current version of ez2boot
// @Tags 			util
// @Success 		200 {object} VersionResponse
// @Router       	/version [get]
// @Security 		BasicAuth
func (h *Handler) GetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, err := h.Service.getVersion()
		if err != nil {
			h.Logger.Error("Failed to get latest version", "domain", "util", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Data: version, Error: "Failed to get latest version"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: version})
	}
}
