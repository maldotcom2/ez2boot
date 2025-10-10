package handler

import (
	"encoding/json"
	"ez2boot/internal/repository"
	"log/slog"
	"net/http"
)

func GetServers(repo *repository.Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := repo.GetServers()
		if err != nil {
			logger.Error("Failed to get servers", "error", err)
			http.Error(w, "Failed to get servers", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(servers)
		if err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
