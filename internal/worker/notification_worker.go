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
				notifications, err := w.NotificationService.GetPendingNotifications()
				if err != nil {
					w.Logger.Error("Error while getting pending notifications", "error", err)
					continue
				}

				if len(notifications) == 0 {
					w.Logger.Debug("No pending notifications")
				}

				for _, n := range notifications {
					sender, ok := w.NotificationService.GetNotificationSender(n.Type)
					if !ok {
						w.Logger.Error("Notification type not supported. Removing from queue", "id", n.Id, "type", n.Type)
						if err := w.NotificationService.DeleteNotification(n.Id); err != nil {
							w.Logger.Error("Could not delete notification from queue", "id", n.Id, "error", err)
							continue
						}
					}

					if err := sender.Send(n.Msg, n.Title, n.Cfg); err != nil {
						w.Logger.Error("Failed to send notification. Removing from queue", "id", n.Id, "error", err)
						if err := w.NotificationService.DeleteNotification(n.Id); err != nil {
							w.Logger.Error("Could not delete notification from queue", "id", n.Id, "error", err)
							continue
						}
					}

					// Delete after successful send
					if err := w.NotificationService.DeleteNotification(n.Id); err != nil {
						w.Logger.Error("Could not delete notification from queue", "id", n.Id, "error", err)
						continue
					}

					w.Logger.Debug("Successfully sent notification and removed from queue", "id", n.Id)
				}
			}
		}
	}()
}
