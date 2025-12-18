package session

import "errors"

func (s *Service) validateServerSession(session ServerSessionRequest) error {
	if session.ServerGroup == "" {
		return errors.New("missing server group ")
	}

	if session.Duration == "" {
		return errors.New("missing duration")
	}

	return nil
}
