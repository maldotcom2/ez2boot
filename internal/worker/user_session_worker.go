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
				if err := w.UserService.ProcessUserSessions(); err != nil {
					w.Logger.Error("Error processing user sessions", "error", err)
				}
			}
		}
	}()
}
