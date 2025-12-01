package email

import (
	"encoding/json"
	"ez2boot/internal/notification"
	"fmt"
	"net/smtp"
)

// Register itself
func init() {
	notification.Register(&EmailNotification{})
}

func (e *EmailNotification) Type() string {
	return "email"
}

func (e *EmailNotification) Label() string {
	return "Email"
}

func (e *EmailNotification) Send(msg string, title string, cfgStr string) error {
	var cfg EmailConfig
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return err
	}
	// message string assembly
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", cfg.To, cfg.From, title, msg)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Unauthenticated mail
	if !cfg.Auth {
		var auth smtp.Auth = nil
		if err := smtp.SendMail(addr, auth, cfg.From, []string{cfg.To}, []byte(message)); err != nil {
			return err
		}

		return nil
	}

	// Authenticated mail
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	if err := smtp.SendMail(addr, auth, cfg.From, []string{cfg.To}, []byte(message)); err != nil {
		return err
	}

	return nil
}
