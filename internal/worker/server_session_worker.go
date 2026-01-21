package worker

import (
	"context"
	"time"
)

// Handle expired or aging server sessions
func StartServerSessionWorker(w Worker, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.Config.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Process expired or aging sessions
				w.SessionService.ProcessServerSessions(ctx)
			}
		}
	}()
}
