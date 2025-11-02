package session

import "errors"

func (s *Service) validateServerSession(session ServerSession) error {
	if session.ServerGroup == "" {
		return errors.New("Missing server group ")
	}

	if session.Duration == "" {
		return errors.New("Missing duration")
	}

	return nil
}
