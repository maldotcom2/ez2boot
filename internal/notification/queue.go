package notification

import (
	"database/sql"
	"ez2boot/internal/shared"
	"time"
)

// Add new notification to queue
func (s *Service) QueueNotification(tx *sql.Tx, n NewNotification) error {
	n.Time = time.Now().Unix()
	if err := s.Repo.queueNotification(tx, n); err != nil {
		return err
	}

	return nil
}

// Get all pending notifications
func (s *Service) GetPendingNotifications() ([]Notification, error) {
	notifications, err := s.Repo.getPendingNotifications()
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// Delete notification by notification ID
func (s *Service) DeleteNotification(id int64) error {
	rows, err := s.Repo.deleteNotificationFromQueue(id)
	if err != nil {
		return err
	}

	if rows == 0 {
		return shared.ErrNoRowsDeleted
	}

	return nil
}
