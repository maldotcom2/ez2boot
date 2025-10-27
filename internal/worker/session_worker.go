package worker

import (
	"context"
	"time"
)

// Handle expired or aging sessions
func StartSessionWorker(w Worker, ctx context.Context) {
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
				expiredSessions, agingSessions, err := w.SessionService.FindExpiredOrAgingSessions()
				if err != nil {
					w.Logger.Error("Error when trying to find aging or expired sessions", "error", err)
					continue
				}

				if len(expiredSessions) == 0 {
					w.Logger.Debug("No expired sessions")
				} else {
					w.SessionService.ProcessExpiredSessions(expiredSessions)
				}

				if len(agingSessions) == 0 {
					w.Logger.Debug("No sessions nearing expiry")
				} else {
					w.SessionService.ProcessAgingSessions(agingSessions)
				}

				// Terminated sessions
				sessionsForCleanup, err := w.SessionService.FindSessionsForAction(1, 1, 1, "off")
				if err != nil {
					w.Logger.Error("Error occurred while finding sessions for cleanup", "error", err)
				}

				if len(sessionsForCleanup) == 0 {
					w.Logger.Debug("No sessions for cleanup")
				} else {
					w.SessionService.CleanupSessions(sessionsForCleanup)
				}

				// Ready-for-use sessions
				sessionsForUse, err := w.SessionService.FindSessionsForAction(0, 0, 0, "on")
				if err != nil {
					w.Logger.Error("Error occurred while finding sessions ready for use", "error", err)
				}

				if len(sessionsForUse) == 0 {
					w.Logger.Debug("No new sessions ready for use")
				} else {
					w.Logger.Debug("New sessions ready for use")
					for _, session := range sessionsForUse {
						// TODO Queue notification
						if err = w.SessionService.SetOnNotifiedFlag(1, session.ServerGroup); err != nil {
							w.Logger.Error("Failed up set flag for session notified on", "error", err)
						}

					}
				}
			}
		}
	}()
}
