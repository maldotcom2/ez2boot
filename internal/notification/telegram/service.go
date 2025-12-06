package telegram

import (
	"bytes"
	"encoding/json"
	"ez2boot/internal/notification"
	"ez2boot/internal/shared"
	"fmt"
	"net/http"
)

// Register itself
func init() {
	notification.Register(&TelegramChannel{})
}

func (s *TelegramChannel) Type() string {
	return "telegram"
}

func (s *TelegramChannel) Label() string {
	return "Telegram"
}

func (s *TelegramChannel) Send(msg string, title string, cfgStr string) error {
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

// Check required Telegram fields.
func (t *TelegramChannel) Validate(config map[string]any) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is missing: %w", shared.ErrFieldMissing)
	}

	chatID, ok := config["chat_id"].(string)
	if !ok || chatID == "" {
		return fmt.Errorf("chat id is missing: %w", shared.ErrFieldMissing)
	}

	return nil
}

// Marshal json
func (t *TelegramChannel) ToConfig(config map[string]any) (string, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
