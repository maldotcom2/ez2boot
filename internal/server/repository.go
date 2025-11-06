package server

import (
	"fmt"
)

// Return all servers from catalogue - names and groups
func (r *Repository) getServers() (map[string][]Server, error) {
	rows, err := r.Base.DB.Query("SELECT unique_id, name, state, server_group FROM servers")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	servers := make(map[string][]Server)
	for rows.Next() {
		var s Server
		err = rows.Scan(&s.UniqueID, &s.Name, &s.State, &s.ServerGroup)
		if err != nil {
			return nil, err
		}
		servers[s.ServerGroup] = append(servers[s.ServerGroup], s)
	}

	return servers, nil
}

func (r *Repository) deleteObsolete(ids []interface{}, placeholderStr string) error {
	query := fmt.Sprintf(`DELETE FROM servers WHERE unique_id NOT IN (%s)`, placeholderStr)
	_, err := r.Base.DB.Exec(query, ids...) // expand
	if err != nil {
		return err
	}

	return nil
}

// Insert new server records, if conflict update the name, server group or state
func (r *Repository) addOrUpdate(server Server) error {
	query := `INSERT INTO servers (unique_id, name, state, server_group, time_added) VALUES ($1, $2, $3, $4, $5) 
			ON CONFLICT (unique_id) DO UPDATE 
			SET name = EXCLUDED.name, state = EXCLUDED.state, server_group = EXCLUDED.server_group
			WHERE servers.name <> EXCLUDED.name OR servers.state <> EXCLUDED.state OR servers.server_group <> EXCLUDED.server_group`

	if _, err := r.Base.DB.Exec(query, server.UniqueID, server.Name, server.State, server.ServerGroup, server.TimeAdded); err != nil {
		return err
	}

	return nil
}

// Get server IDs which are pending a state change
func (r *Repository) getPending(currentState string, nextState string) ([]string, error) {
	rows, err := r.Base.DB.Query(`SELECT unique_id FROM servers WHERE state = $1 AND next_state = $2`, currentState, nextState)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	serverIDs := []string{}

	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, err
		}

		serverIDs = append(serverIDs, s)
	}

	return serverIDs, nil
}
