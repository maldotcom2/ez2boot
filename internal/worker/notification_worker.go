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
				// Break out of Go Routine
				return
			case <-ticker.C:
				// Get pending notifications
				if err := w.NotificationService.ProcessNotifications(); err != nil {
					w.Logger.Error("Error while processing notifications", "error", err)
				}
			}
		}
	}()
}
