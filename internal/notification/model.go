package notification

type Sender interface {
	Type() string                                      // Get the name
	Send(cfg string, msg string, subject string) error // Send the notification
}
