package telegram

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"fmt"
)

// Check required Telegram fields.
func (t *TelegramHandler) Validate(config map[string]any) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is missing: %w", shared.ErrFieldMissing)
	}

	chatID, ok := config["chatId"].(string)
	if !ok || chatID == "" {
		return fmt.Errorf("chat id is missing: %w", shared.ErrFieldMissing)
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
