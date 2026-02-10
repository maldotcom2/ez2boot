package teams

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
	notification.Register(&TeamsChannel{})
}

func (e *TeamsChannel) Type() string {
	return "teams"
}

func (e *TeamsChannel) Label() string {
	return "Teams"
}

func (s *TeamsChannel) Send(msg string, title string, cfgStr string) error {
	var cfg TeamsConfig
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return err
	}

	payload := map[string]any{
		"type":    "AdaptiveCard",
		"version": "1.4",
		"body": []map[string]any{
			{
				"type": "TextBlock",
				"text": msg,
				"wrap": true,
			},
		},
	}

	body, _ := json.Marshal(payload)

	_, err := http.Post(
		cfg.Webhook,
		"application/json",
		bytes.NewReader(body),
	)

	return err
}

// Check required Teams fields.
func (t *TeamsChannel) Validate(config map[string]any) error {
	webhook, ok := config["webhook"].(string)
	if !ok || webhook == "" {
		return fmt.Errorf("webhook is missing: %w", shared.ErrFieldMissing)
	}

	return nil
}

// Marshal json
func (t *TeamsChannel) ToConfig(config map[string]any) (string, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
