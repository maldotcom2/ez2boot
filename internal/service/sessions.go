package service

import (
	"context"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

// Handle expired or aging sessions
func StartSessionWorker(repo *repository.Repository, ctx context.Context, cfg model.Config, logger *slog.Logger) {
	go func() {
		ticker := time.NewTicker(cfg.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Process expired or aging sessions
				expiredSessions, agingSessions, err := findExpiredOrAgingSessions(repo)
				if err != nil {
					logger.Error("Error when trying to find aging or expired sessions", "error", err)
					continue
				}

				if len(expiredSessions) == 0 {
					logger.Debug("No expired sessions")
				} else {
					processExpiredSessions(repo, expiredSessions, logger)
				}

				if len(agingSessions) == 0 {
					logger.Debug("No sessions nearing expiry")
				} else {
					processAgingSessions(repo, agingSessions, logger)
				}

				// Terminated sessions
				sessionsForCleanup, err := findSessionsForAction(repo, 1, 1, "off")
				if err != nil {
					logger.Error("Error occurred while finding sessions for cleanup", "error", err)
				}

				if len(sessionsForCleanup) == 0 {
					logger.Debug("No sessions for cleanup")
				} else {
					repo.CleanupSessions(sessionsForCleanup)
				}

				// Ready-for-use sessions
				sessionsForUse, err := findSessionsForAction(repo, 0, 0, "on")
				if err != nil {
					logger.Error("Error occurred while finding sessions ready for use", "error", err)
				}

				if len(sessionsForUse) == 0 {
					logger.Debug("No new sessions ready for use")
				} else {
					logger.Debug("New sessions ready for use")
					for _, session := range sessionsForUse {
						// TODO Queue notification
						if err = repo.SetOnNotifiedFlag(1, session.ServerGroup); err != nil {
							logger.Error("Failed up set flag for session notified on", "error", err)
						}

					}
				}
			}
		}
	}()
}

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
