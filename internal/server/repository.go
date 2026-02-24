package server

import (
	"fmt"
	"strings"
)

func (r *Repository) deleteObsolete(ids []any) error {
	// Successful scrape returned nothing, means remove all
	if len(ids) == 0 {
		_, err := r.Base.DB.Exec("DELETE FROM servers")
		return err
	}

	// Build string of positional placeholders eg $1, $2, $3
	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("DELETE FROM servers WHERE unique_id NOT IN (%s)", strings.Join(placeholders, ", "))
	if _, err := r.Base.DB.Exec(query, ids...); err != nil {
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
	rows, err := r.Base.DB.Query("SELECT unique_id FROM servers WHERE state = $1 AND next_state = $2", currentState, nextState)
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
