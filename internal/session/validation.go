package session

import (
	"ez2boot/internal/shared"
	"time"
)

func (s *Service) validateServerSession(session ServerSessionRequest) error {
	if session.ServerGroup == "" {
		return shared.ErrFieldMissing
	}

	if session.Duration == "" {
		return shared.ErrFieldMissing
	}

	dur, err := time.ParseDuration(session.Duration)
	if err != nil {
		return err
	}

	// Check if duration is beyond max
	if dur > s.Config.MaxServerSessionDuration {
		return shared.ErrDurationTooLong
	}

	return nil
}
