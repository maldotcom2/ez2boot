package session

import (
	"errors"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"time"
)

func (s *Service) getServerSessions() ([]ServerSession, error) {
	sessions, err := s.Repo.getServerSessions()
	if err != nil {
		return []ServerSession{}, err
	}

	return sessions, nil
}

func (s *Service) updateServerSession(session ServerSession) (ServerSession, error) {
	if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
		return ServerSession{}, errors.New("email, server_group, duration is required")
	}

	updated, updatedSession, err := s.Repo.updateServerSession(session)
	if err != nil {
		return ServerSession{}, err
	}

	if !updated {
		return ServerSession{}, shared.ErrSessionNotFound
	}

	return updatedSession, nil
}

func (s *Service) newServerSession(session ServerSession) (ServerSession, error) {
	if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
		return ServerSession{}, errors.New("Email, server_group and duration required")
	}

	// Generate token
	token, err := util.GenerateRandomString(16)
	if err != nil {
		return ServerSession{}, err
	}

	session.Token = token
	session, err = s.Repo.newServerSession(session)
	if err != nil {
		return ServerSession{}, err
	}

	return session, nil
}

func (s *Service) CleanupServerSessions(sessions []ServerSession) {
	s.Repo.cleanupServerSessions(sessions)
}

// Find expired sessions
func (s *Service) FindExpiredOrAgingServerSessions() ([]ServerSession, []ServerSession, error) {
	currentSessions, err := s.Repo.getServerSessions()
	if err != nil {
		return nil, nil, err
	}

	var expiredSessions []ServerSession
	var agingSessions []ServerSession
	now := time.Now().UTC()
	warningWindow := now.Add(15 * time.Minute) //TODO make adjustable

	for _, session := range currentSessions {
		if session.Expiry.Before(now) {
			expiredSessions = append(expiredSessions, session)
		} else if session.Expiry.Before(warningWindow) {
			agingSessions = append(agingSessions, session)
		}
	}

	return expiredSessions, agingSessions, nil
}

func (s *Service) ProcessExpiredSessions(expiredSessions []ServerSession) {
	s.Logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		if err := s.Repo.endServerSession(session.ServerGroup); err != nil {
			s.Logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) ProcessAgingSessions(agingSessions []ServerSession) {
	s.Logger.Debug("Found aging sessions", "count", len(agingSessions))

	for _, session := range agingSessions {
		// TODO Queue notification
		if err := s.Repo.setWarningNotifiedFlag(1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) FindServerSessionsForAction(toCleanup int, onNotified int, offNotified int, serverState string) ([]ServerSession, error) {
	sessions, err := s.Repo.findServerSessionsForAction(toCleanup, onNotified, offNotified, serverState)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *Service) SetWarningNotifiedFlag(value int, serverGroup string) error {
	if err := s.Repo.setWarningNotifiedFlag(value, serverGroup); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetOnNotifiedFlag(value int, serverGroup string) error {
	if err := s.Repo.setOnNotifiedFlag(value, serverGroup); err != nil {
		return err
	}

	return nil
}
