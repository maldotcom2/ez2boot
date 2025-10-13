package service

import (
	"context"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

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
				expiredSessions, err := findExpiredSessions(repo)
				if err != nil {
					logger.Error("Failed to find expired sessions", "error", err)
					continue
				}

				if len(expiredSessions) == 0 {
					logger.Debug("No expired sessions found")
					continue
				}

				logger.Info("Found expired sessions", "count", len(expiredSessions))
				for _, session := range expiredSessions {
					if err := repo.EndSession(session.ServerGroup); err != nil {
						logger.Error("Failed to cleanup expired session", "error", err)
					} else {
						logger.Info("Ended session, notify pending", "email", session.Email)
					}
				}
			}
		}
	}()
}

// Find expired sessions
func findExpiredSessions(repo *repository.Repository) ([]model.Session, error) {
	currentSessions, err := repo.GetSessions()
	if err != nil {
		return []model.Session{}, err
	}

	var expiredSessions []model.Session
	now := time.Now().UTC()

	for _, session := range currentSessions {
		if session.Expiry.Before(now) {
			expiredSessions = append(expiredSessions, session)
		}
	}

	return expiredSessions, nil
}
