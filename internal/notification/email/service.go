package email

import (
	"encoding/json"
	"ez2boot/internal/notification"
	"ez2boot/internal/shared"
	"fmt"
	"net/smtp"
)

// Register itself
func init() {
	notification.Register(&EmailChannel{})
}

func (e *EmailChannel) Type() string {
	return "email"
}

func (e *EmailChannel) Label() string {
	return "Email"
}

func (e *EmailChannel) Send(msg string, title string, cfgStr string) error {
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

// Checks required Email fields.
func (e *EmailChannel) Validate(cfg map[string]any) error {
	host, ok := cfg["host"].(string)
	if !ok || host == "" {
		return fmt.Errorf("host is missing: %w", shared.ErrFieldMissing)
	}

	port, ok := cfg["port"].(float64)
	if !ok || port == 0 {
		return fmt.Errorf("port is missing - must be an integer: %w", shared.ErrFieldMissing)
	}

	to, ok := cfg["to"].(string)
	if !ok || to == "" {
		return fmt.Errorf("to is missing: %w", shared.ErrFieldMissing)
	}

	from, ok := cfg["from"].(string)
	if !ok || from == "" {
		return fmt.Errorf("from is missing: %w", shared.ErrFieldMissing)
	}

	// Auth is optional, but user and password are required
	auth, _ := cfg["auth"].(bool)
	if auth {
		user, _ := cfg["user"].(string)
		password, _ := cfg["password"].(string)
		if user == "" || password == "" {
			return shared.ErrMissingAuthValues
		}
	}

	return nil
}

// Marshal to json
func (e *EmailChannel) ToConfig(config map[string]any) (string, error) {
	// Remove credentials if auth is false
	auth, _ := config["auth"].(bool)
	if !auth {
		config["user"] = ""
		config["password"] = ""
	}

	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
