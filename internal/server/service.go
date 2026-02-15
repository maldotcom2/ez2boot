package server

// Update servers from cloud provider
func (s *Service) UpdateServers(servers []Server) {
	// Extract UniqueIDs into a slice of interface{}
	ids := make([]any, len(servers))
	for i, s := range servers {
		ids[i] = s.UniqueID
	}

	// Delete servers from DB not in scrape
	err := s.Repo.deleteObsolete(ids)
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

// Get server IDs which are pending a state change
func (s *Service) GetPending(currentState string, nextState string) ([]string, error) {
	serverIDs, err := s.Repo.getPending(currentState, nextState)
	if err != nil {
		return nil, err
	}

	return serverIDs, nil
}
