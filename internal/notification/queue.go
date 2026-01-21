package notification

import (
	"context"
	"database/sql"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"time"
)

func (s *Service) ProcessNotifications(ctx context.Context) error {
	// Remove any from queue where user does not have configured notifications
	rows, err := s.Repo.deleteOrphanedNotifications()
	if err != nil {
		s.Logger.Error("Error while removing orphaned notifications", "error", err)
		return err
	}

	if rows > 0 {
		s.Logger.Debug("Deleted orphaned notifications", "count", rows)
	}

	// Get remaining pending
	notifications, err := s.Repo.getPendingNotifications()
	if err != nil {
		s.Logger.Error("Error while getting pending notifications", "error", err)
		return err
	}

	// Nothing to do
	if len(notifications) == 0 {
		s.Logger.Debug("No pending notifications")
		return nil
	}

	// Get sender for each
	for _, n := range notifications {
		sender, ok := s.getNotificationSender(n.Type)
		if !ok {
			s.Logger.Error("Notification type not supported. Removing from queue", "id", n.Id, "type", n.Type, "title", n.Title)
			_, err := s.Repo.deleteNotificationFromQueue(n.Id)
			if err != nil {
				s.Logger.Error("Could not delete notification from queue", "id", n.Id, "type", n.Type, "title", n.Title, "error", err)
			}
			continue
		}

		// Send
		sent := false
		for i := 0; i < 3; i++ {
			if err := sender.Send(n.Msg, n.Title, n.Cfg); err != nil {
				s.Logger.Error("Failed to send notification", "attempt", i+1, "id", n.Id, "type", n.Type, "title", n.Title, "error", err)
				continue
			}
			// success
			sent = true

			actorUserID, actorEmail := ctxutil.GetActor(ctx)
			s.Audit.Log(audit.Event{
				ActorUserID:  actorUserID,
				ActorEmail:   actorEmail,
				TargetUserID: n.UserID,
				Action:       "sent",
				Resource:     "notification",
				Success:      true,
				Metadata: map[string]any{
					"type":  n.Type,
					"title": n.Title,
				},
			})

			break
		}

		// Delete notification whether it was sent successfully or not
		_, err := s.Repo.deleteNotificationFromQueue(n.Id)
		if err != nil {
			s.Logger.Error("Could not delete notification from queue", "id", n.Id, "error", err)
			continue
		}

		if sent {
			s.Logger.Debug("Successfully sent notification and removed from queue", "id", n.Id, "type", n.Type, "title", n.Title)
		} else {
			s.Logger.Warn("Failed to send notification", "id", n.Id, "type", n.Type, "title", n.Title)
		}
	}

	return nil
}

// Add new notification to queue
func (s *Service) QueueNotification(tx *sql.Tx, n NewNotification) error {
	n.Time = time.Now().Unix()
	if err := s.Repo.queueNotification(tx, n); err != nil {
		return err
	}

	return nil
}
