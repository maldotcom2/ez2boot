package session

import (
	"errors"
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"ez2boot/internal/util"
	"log/slog"
	"time"
)

func (s *Service) createNewSession(session model.Session) (model.Session, error) {
	if session.Email == "" || session.ServerGroup == "" || session.Duration == "" {
		return model.Session{}, errors.New("Email, server_group and duration required")
	}

	// Generate token
	token, err := util.GenerateToken(16)
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

// Find expired sessions
func (s *Service) findExpiredOrAgingSessions(repo *db.Repository) ([]model.Session, []model.Session, error) {
	currentSessions, err := s.Repo.GetAllSessions()
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

func (s *Service) processExpiredSessions(repo *db.Repository, expiredSessions []model.Session, logger *slog.Logger) {
	logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		if err := s.Repo.EndSession(session.ServerGroup); err != nil {
			logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) processAgingSessions(repo *db.Repository, agingSessions []model.Session, logger *slog.Logger) {
	logger.Debug("Found aging sessions", "count", len(agingSessions))

	for _, session := range agingSessions {
		// TODO Queue notification
		if err := s.Repo.SetWarningNotifiedFlag(1, session.ServerGroup); err != nil {
			logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func (s *Service) findSessionsForAction(repo *db.Repository, toCleanup int, onNotified int, serverState string) ([]model.Session, error) {
	sessions, err := s.Repo.FindSessionsForAction(toCleanup, onNotified, serverState)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
