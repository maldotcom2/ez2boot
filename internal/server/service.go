package server

import "ez2boot/internal/model"

func (s *Service) GetServers() (map[string][]model.Server, error) {
	servers, err := s.Repo.GetServers()
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (s *Service) UpdateServers(servers []model.Server) {
	s.Repo.UpdateServers(servers)
}
