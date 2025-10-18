package service

import (
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

// Find expired sessions
func findExpiredOrAgingSessions(repo *repository.Repository) ([]model.Session, []model.Session, error) {
	currentSessions, err := repo.GetSessions()
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

func processExpiredSessions(repo *repository.Repository, expiredSessions []model.Session, logger *slog.Logger) {
	logger.Debug("Found expired sessions", "count", len(expiredSessions))

	for _, session := range expiredSessions {
		if err := repo.EndSession(session.ServerGroup); err != nil {
			logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func processAgingSessions(repo *repository.Repository, agingSessions []model.Session, logger *slog.Logger) {
	logger.Debug("Found aging sessions", "count", len(agingSessions))

	for _, session := range agingSessions {
		// TODO Queue notification
		if err := repo.SetWarningNotifiedFlag(1, session.ServerGroup); err != nil {
			logger.Error("Failed to set session as notified", "email", session.Email, "server_group", session.ServerGroup, "error", err)
		}
	}
}

func findSessionsForAction(repo *repository.Repository, toCleanup int, onNotified int, serverState string) ([]model.Session, error) {
	sessions, err := repo.FindSessionsForAction(toCleanup, onNotified, serverState)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
