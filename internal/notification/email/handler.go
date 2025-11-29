package email

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"fmt"
)

// Checks required Email fields.
func (e *EmailHandler) Validate(cfg map[string]any) error {
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
func (e *EmailHandler) ToConfig(config map[string]any) (string, error) {
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
