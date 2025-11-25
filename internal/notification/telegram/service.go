package telegram

import (
	"bytes"
	"encoding/json"
	"ez2boot/internal/notification"
	"net/http"
)

// Register itself
func init() {
	notification.Register(&TelegramNotification{})
}

func (s *TelegramNotification) Type() string {
	return "telegram"
}

func (s *TelegramNotification) Label() string {
	return "Telegram"
}

func (s *TelegramNotification) Send(msg string, title string, cfgStr string) error {
	var cfg TelegramConfig
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"chat_id": cfg.ChatID,
		"text":    msg,
	}

	body, _ := json.Marshal(payload)

	_, err := http.Post(
		"https://api.telegram.org/bot"+cfg.Token+"/sendMessage",
		"application/json",
		bytes.NewReader(body),
	)

	return err
}

func (s *Service) AddOrUpdate(userID int64, cfg TelegramConfig) error {
	// Must supply
	if cfg.Token == "" || cfg.ChatID == "" {
		return ErrMissingValues
	}

	// Stringify the config
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	jsonStr := string(data)

	if err = s.Repo.addOrUpdate(userID, jsonStr); err != nil {
		return err
	}

	return nil
}
