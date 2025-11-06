package worker

import (
	"context"
	"time"
)

// Cleanup expired user login sessions
func StartExpiredUserSessionCleanup(w Worker, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.Config.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Process expired user sessions
				result, err := w.UserService.DeleteExpiredUserSessions()
				if err != nil {
					w.Logger.Error("Error while deleting expired user sessions", "error", err)
					continue
				}

				if result == nil {
					w.Logger.Debug("No expired user sessions to cleanup")
					continue
				}

				rows, err := result.RowsAffected()
				if err != nil {
					w.Logger.Error("Error getting affected rows for user session cleanup", "error", err)
					continue
				}

				if rows > 0 {
					w.Logger.Debug("Deleted expired sessions", "count", rows)
				}
			}
		}
	}()
}
