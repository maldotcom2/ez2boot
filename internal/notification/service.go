package notification

// Global in-memory store of available notification channels // TODO is this actually needed? Will the FE validate?
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
func (s *Service) getNotificationTypes() []string {
	types := make([]string, 0, len(registry))
	for k := range registry {
		types = append(types, k)
	}
	return types
}
