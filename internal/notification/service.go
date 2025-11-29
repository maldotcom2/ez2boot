package notification

import "ez2boot/internal/shared"

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

// Add or update personal notification options
func (s *Service) setUserNotification(userID int64, req NotificationUpdateRequest) error {
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
	return s.Repo.setUserNotification(userID, req.Type, cfgStr)
}
