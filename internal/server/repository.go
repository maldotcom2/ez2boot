package server

import (
	"ez2boot/internal/model"
	"fmt"
	"strings"
)

// Return all servers from catalogue - names and groups
func (r *Repository) GetAllServers() (map[string][]model.Server, error) {
	rows, err := r.Base.DB.Query("SELECT unique_id, name, state, server_group FROM servers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servers := make(map[string][]model.Server)
	for rows.Next() {
		var s model.Server
		err = rows.Scan(&s.UniqueID, &s.Name, &s.State, &s.ServerGroup)
		if err != nil {
			return nil, err
		}
		servers[s.ServerGroup] = append(servers[s.ServerGroup], s)
	}

	return servers, nil
}

// Add or update servers. Errors are not returned here due to GO routine
func (r *Repository) UpdateServers(servers []model.Server) {

	err := r.deleteObsolete(servers)
	if err != nil {
		r.Base.Logger.Error("Failed to delete obsolete servers from local DB", "error", err)
		// Continue
	}

	r.addOrUpdate(servers)
}

func (r *Repository) deleteObsolete(servers []model.Server) error {
	// Extract UniqueIDs into a slice of interface{}
	ids := make([]interface{}, len(servers))
	for i, s := range servers {
		ids[i] = s.UniqueID
	}

	// Build placeholders
	placeholders := make([]string, len(servers))
	for i := range servers {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholderStr := strings.Join(placeholders, ", ")

	// Build the full query with expanded placeholders
	query := fmt.Sprintf(`DELETE FROM servers WHERE unique_id NOT IN (%s)`, placeholderStr)

	_, err := r.Base.DB.Exec(query, ids...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) addOrUpdate(servers []model.Server) {
	const updateQuery = `INSERT INTO servers (unique_id, name, state, server_group, time_added) VALUES ($1, $2, $3, $4, $5) 
						ON CONFLICT (unique_id, name) DO UPDATE 
						SET state = EXCLUDED.state, server_group = EXCLUDED.server_group
						WHERE servers.state IS NOT EXCLUDED.state OR servers.server_group IS NOT EXCLUDED.state`

	for _, server := range servers {
		_, err := r.Base.DB.Exec(updateQuery, server.UniqueID, server.Name, server.State, server.ServerGroup, server.TimeAdded)
		if err != nil {
			r.Base.Logger.Error("Failed to add or update server from scrap", "server", server, "error", err) // Log here to show error and continue
		}
	}
}
