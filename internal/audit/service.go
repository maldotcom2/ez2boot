package audit

func (s *Service) Log(e Event) {
	if err := s.Repo.Log(e); err != nil {
		s.Logger.Error("Failed to write audit log", "error", err)
	}
}
