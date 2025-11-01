package notification

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

func (s *Service) GetPendingNotifications() ([]Notification, error) {
	notifications, err := s.Repo.findPendingNotifications()
	if err != nil {
		return nil, err
	}

	return notifications, nil
}
