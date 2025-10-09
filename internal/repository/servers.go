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

// Add or update servers
func (r *Repository) AddOrUpdateServers(servers []model.Server, logger *slog.Logger) {
	query := `INSERT INTO servers (unique_id, name, state, server_group, time_added) VALUES ($1, $2, $3, $4, $5) 
			ON CONFLICT (unique_id, name) DO UPDATE 
			SET state = EXCLUDED.state, server_group = EXCLUDED.server_group
			WHERE servers.state IS NOT EXCLUDED.state OR servers.server_group IS NOT EXCLUDED.state`

	// TODO remove from DB any servers which did not appear in the scrape
	for _, server := range servers {
		_, err := r.DB.Exec(query, server.UniqueID, server.Name, server.State, server.ServerGroup, server.TimeAdded)
		if err != nil {
			logger.Error("Failed to insert or update status for server:", "name", server.Name, "err", err)
		}
	}
}
