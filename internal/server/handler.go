package server

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := h.Service.GetServers()
		if err != nil {
			h.Logger.Error("Failed to get servers", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get servers"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: servers})
	}
}
