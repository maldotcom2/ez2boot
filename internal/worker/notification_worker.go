package worker

import (
	"context"
	"ez2boot/internal/notification"
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
				notifications, err := w.NotificationService.GetPendingNotifications()
				if err != nil {
					w.Logger.Error("Error while getting pending notifications", "error", err)
					continue
				}

				if len(notifications) == 0 {
					w.Logger.Debug("No pending notifications")
				}

				for _, n := range notifications {
					sender, ok := notification.GetSender(n.Type)
					if !ok {
						w.Logger.Error("Notification type not supported", "type", n.Type)
						// TODO delete this from table
						continue
					}

					if err := sender.Send(n.Msg, n.Title, n.Cfg); err != nil {
						w.Logger.Error("Failed to send notification", "error", err)
						// TODO delete from table
						continue
					}
				}
			}
		}
	}()
}
