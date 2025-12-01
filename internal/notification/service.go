package notification

import (
	"database/sql"
	"errors"
	"ez2boot/internal/shared"
)

// Global in-memory store of available notification channels
var registry = make(map[string]Sender)

// Add sender to registry - notification packages register via their inits when imported
func Register(sender Sender) {
	registry[sender.Type()] = sender
}

// Retrieves sender by type name, return value can then be called for sending notification, eg sender, ok := GetSender("email"). sender.Send(params)
// Used by notification worker
func (s *Service) getNotificationSender(typeName string) (Sender, bool) {
	sender, ok := registry[typeName]
	return sender, ok
}

// Retrieves all supported notification types
func (s *Service) getNotificationTypes() []NotificationTypeRequest {
	list := make([]NotificationTypeRequest, 0, len(registry))
	for _, sender := range registry {
		list = append(list, NotificationTypeRequest{
			Type:  sender.Type(),
			Label: sender.Label(),
		})
	}
	return list
}

// Get current user notification settings
func (s *Service) getUserNotificationSettings(userID int64) (NotificationConfigResponse, error) {
	raw, err := s.Repo.getUserNotificationSettings(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User hasn't configured notifications yet
			return NotificationConfigResponse{}, nil
		}

		return NotificationConfigResponse{}, err
	}

	cc := raw.ChannelConfig
	hasPassword := false

	// Check for sensitive value
	pw, ok := cc["password"].(string)
	if ok && pw != "" {
		hasPassword = true
		delete(cc, "password")
	}

	cc["has_password"] = hasPassword

	return NotificationConfigResponse{
		Type:          raw.Type,
		ChannelConfig: cc,
	}, nil
}

// Add or update personal notification options
func (s *Service) setUserNotificationSettings(userID int64, req NotificationConfigRequest) error {
	// Check the notification type is supported
	handler, ok := s.Handlers[req.Type]
	if !ok {
		return shared.ErrNotificationTypeNotSupported
	}

	// Call handler specific validation
	if err := handler.Validate(req.ChannelConfig); err != nil {
		return err
	}

	// Call handler specific marshaler
	cfgStr, err := handler.ToConfig(req.ChannelConfig)
	if err != nil {
		return err
	}

	// Store it
	if err := s.Repo.setUserNotificationSettings(userID, req.Type, cfgStr); err != nil {
		return err
	}

	return nil
}

func (s *Service) deleteUserNotificationSettings(userID int64) error {
	if err := s.Repo.deleteUserNotificationSettings(userID); err != nil {
		return err
	}

	return nil
}
