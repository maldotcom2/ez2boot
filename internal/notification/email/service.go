package email

import (
	"encoding/json"
	"ez2boot/internal/notification"
	"fmt"
	"net/smtp"
)

func init() {
	notification.Register(&EmailNotification{})
}

func (e *EmailNotification) Type() string {
	return "email"
}

func (e *Service) AddOrUpdate(userID int64, cfg EmailConfig) error {
	// Must supply creds for auth
	if cfg.Auth && (cfg.User == "" || cfg.Password == "") {
		return ErrMissingAuthValues
	}

	// Stringify the config
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	jsonStr := string(data)

	if err = e.Repo.addOrUpdate(userID, jsonStr); err != nil {
		return err
	}

	return nil
}

func (e *EmailNotification) Send(msg string, title string, cfgStr string) error {

	var cfg EmailConfig
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return err
	}
	// message string assembly
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", cfg.To, cfg.From, title, msg)

	// TODO option to set nil for unauthenticated
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	// Send the email
	err := smtp.SendMail(cfg.Host+":"+cfg.Port, auth, cfg.From, []string{cfg.To}, []byte(message))
	if err != nil {
		return err
	}

	return nil
}
