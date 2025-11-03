package server

import (
	"fmt"
	"strings"
)

// Returns map of servers for each server group
func (s *Service) GetServers() (map[string][]Server, error) {
	servers, err := s.Repo.GetServers()
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// Update servers from cloud provider
func (s *Service) UpdateServers(servers []Server) {
	if len(servers) == 0 {
		s.Logger.Warn("No servers to update")
		return
	}

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

	// Delete servers from DB not in scrape
	err := s.Repo.deleteObsolete(ids, placeholderStr)
	if err != nil {
		s.Logger.Error("Failed to delete obsolete servers from DB", "error", err)
	}

	// Process update
	for _, server := range servers {
		if err := s.Repo.addOrUpdate(server); err != nil {
			s.Logger.Error("Failed to add or update server from scrape", "server", server, "error", err) // Log here to show error and continue
			continue
		}
	}
}
