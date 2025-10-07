package repository

import (
	"ez2boot/internal/model"
	"log/slog"
)

// Return all servers from catalogue - names and groups
func (r *Repository) GetServers(logger *slog.Logger) ([]model.Server, error) {
	rows, err := r.DB.Query("SELECT name, server_group FROM servers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servers := []model.Server{}
	for rows.Next() {
		var s model.Server
		err = rows.Scan(&s.Name, &s.ServerGroup)
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}
