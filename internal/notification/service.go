package notification

import (
	"database/sql"
	"ez2boot/internal/shared"
	"time"
)

// In memory store of available notification channels
var registry = map[string]Sender{}

// Add sender to registry - notification packages register via their inits when imported
func Register(sender Sender) {
	registry[sender.Type()] = sender
}

// Retrieves sender by type name, return value can then be called for sending notification, eg sender, ok := GetSender("email"). sender.Send(params)
func GetSender(typeName string) (Sender, bool) {
	s, ok := registry[typeName]
	return s, ok
}

// Retrieves all supported notification types - called externally
func SupportedTypes() []string {
	types := make([]string, 0, len(registry))
	for k := range registry {
		types = append(types, k)
	}
	return types
}

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
