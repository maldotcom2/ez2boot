package telegram

import (
	"encoding/json"
)

// Check required Telegram fields.
func (t *TelegramHandler) Validate(config map[string]any) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return ErrMissingValues
	}

	chatID, ok := config["chatId"].(string)
	if !ok || chatID == "" {
		return ErrMissingValues
	}

	return nil
}

// Marshal json
func (t *TelegramHandler) ToConfig(config map[string]any) (string, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
