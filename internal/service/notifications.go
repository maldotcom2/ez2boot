package service

import (
	"context"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"
)

func CleanupAndNotify(repo *repository.Repository, ctx context.Context, cfg model.Config, logger *slog.Logger) error {

	// TODO notification channel selector

	StartCleanupAndNotificationWorker(repo, ctx, cfg, logger)

	return nil
}

func StartCleanupAndNotificationWorker(repo *repository.Repository, ctx context.Context, cfg model.Config, logger *slog.Logger) {
	go func() {
		ticker := time.NewTicker(cfg.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				sessionsForNotify, err := repo.CleanupSessions()
				if err != nil {
					logger.Error("Failed to run session cleanup", "error", err)
				}

				if len(sessionsForNotify) == 0 {
					// TODO logger
					continue
				}

				for _, session := range sessionsForNotify {
					// TODO send Notification
					logger.Error("Failed to send notification", "session", session.ServerGroup, "error", err)
				}
			}
		}
	}()
}
