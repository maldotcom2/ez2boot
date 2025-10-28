package server

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) GetServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := h.Service.GetServers()
		if err != nil {
			h.Logger.Error("Failed to get servers", "error", err)
			http.Error(w, "Failed to get servers", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(servers)
		if err != nil {
			h.Logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
