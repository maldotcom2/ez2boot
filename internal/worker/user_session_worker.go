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
				w.Logger.Debug("Exiting user session worker", "domain", "worker")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running user session worker", "domain", "worker")
				// Process expired user sessions
				if err := w.UserService.ProcessUserSessions(); err != nil {
					w.Logger.Error("Failed to process user sessions", "domain", "worker", "error", err)
				}
			}
		}
	}()
}
