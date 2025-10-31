package notification

// In memory stare of available notification channels
var registry = map[string]Sender{}

// Add sender to registry - called internally by package inits
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
