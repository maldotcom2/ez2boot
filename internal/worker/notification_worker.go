package worker

import (
	"context"
	"time"
)

func StartNotificationWorker(w Worker, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.Config.InternalClock)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.Logger.Debug("Exiting notification worker", "domain", "worker")
				// Break out of Go Routine
				return
			case <-ticker.C:
				w.Logger.Debug("Running notification worker", "domain", "worker")
				// Get pending notifications
				if err := w.NotificationService.ProcessNotifications(ctx); err != nil {
					w.Logger.Error("Failed while processing notifications", "domain", "worker", "error", err)
				}
			}
		}
	}()
}
