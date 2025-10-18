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
