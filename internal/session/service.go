package session

import (
	"errors"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"time"
)

func (s *Service) GetSessions() ([]model.Session, error) {
	sessions, err := s.Repo.GetSessions()
	if err != nil {
		return []model.Session{}, err
	}

	return sessions, nil
}

func (s *Service) UpdateSession(session model.Session) (model.Session, error) {
	if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
		return model.Session{}, errors.New("email, server_group, duration is required")
	}

	updated, updatedSession, err := s.Repo.UpdateSession(session)
	if err != nil {
		return model.Session{}, err
	}

	if !updated {
		return model.Session{}, shared.ErrSessionNotFound
	}

	return updatedSession, nil
}

func (s *Service) createNewSession(session model.Session) (model.Session, error) {
	if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
		return model.Session{}, errors.New("Email, server_group and duration required")
	}

	// Generate token
	token, err := util.GenerateRandomString(16)
	if err != nil {
		return model.Session{}, err
	}

	session.Token = token
	session, err = s.Repo.NewSession(session)
	if err != nil {
		return model.Session{}, err
	}

	return session, nil
}

func (s *Service) CleanupSessions(sessions []model.Session) {
	s.Repo.CleanupSessions(sessions)
}

// Find expired sessions
func (s *Service) FindExpiredOrAgingSessions() ([]model.Session, []model.Session, error) {
	currentSessions, err := s.Repo.GetSessions()
	if err != nil {
		return nil, nil, err
	}

	var expiredSessions []model.Session
	var agingSessions []model.Session
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

func (s *Service) ProcessExpiredSessions(expiredSessions []model.Session) {
	s.Logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		if err := s.Repo.EndSession(session.ServerGroup); err != nil {
			s.Logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) ProcessAgingSessions(agingSessions []model.Session) {
	s.Logger.Debug("Found aging sessions", "count", len(agingSessions))

	for _, session := range agingSessions {
		// TODO Queue notification
		if err := s.Repo.SetWarningNotifiedFlag(1, session.ServerGroup); err != nil {
			s.Logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) FindSessionsForAction(toCleanup int, onNotified int, offNotified int, serverState string) ([]model.Session, error) {
	sessions, err := s.Repo.FindSessionsForAction(toCleanup, onNotified, offNotified, serverState)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *Service) SetWarningNotifiedFlag(value int, serverGroup string) error {
	if err := s.Repo.SetWarningNotifiedFlag(value, serverGroup); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetOnNotifiedFlag(value int, serverGroup string) error {
	if err := s.Repo.SetOnNotifiedFlag(value, serverGroup); err != nil {
		return err
	}

	return nil
}
