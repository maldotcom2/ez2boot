package server

func (s *Service) GetServers() (map[string][]Server, error) {
	servers, err := s.Repo.GetServers()
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (s *Service) UpdateServers(servers []Server) {
	s.Repo.UpdateServers(servers)
}
