package notification

import (
	"database/sql"
	"encoding/json"
	"errors"
	"ez2boot/internal/encryption"
	"ez2boot/internal/shared"
	"sort"
)

// Global in-memory store of available notification channels
var registry = make(map[string]NotificationChannel)

// Add sender to registry - notification packages register via their inits when imported
func Register(sender NotificationChannel) {
	registry[sender.Type()] = sender
}

// Retrieves sender by type name, return value can then be called for sending notification, eg sender, ok := GetSender("email"). sender.Send(params)
// Used by notification worker
func (s *Service) getNotificationSender(typeName string) (NotificationChannel, bool) {
	sender, ok := registry[typeName]
	return sender, ok
}

// Retrieves all supported notification types
func (s *Service) getNotificationTypes() []NotificationTypeResponse {
	list := make([]NotificationTypeResponse, 0, len(registry))
	for _, sender := range registry {
		list = append(list, NotificationTypeResponse{
			Type:  sender.Type(),
			Label: sender.Label(),
		})
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Type < list[j].Type
	})

	return list
}

// Get current user notification settings
func (s *Service) getUserNotificationSettings(userID int64) (NotificationConfigResponse, error) {
	raw, err := s.Repo.getUserNotificationSettings(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User hasn't configured notifications yet
			return NotificationConfigResponse{}, nil
		}

		return NotificationConfigResponse{}, err
	}

	var resp NotificationConfigResponse
	resp.Type = raw.Type

	// Decrypt notification config
	cfgBytes, err := s.Encryptor.Decrypt([]byte(raw.EncConfig))
	if err != nil {
		return NotificationConfigResponse{}, err
	}

	// Unmarshal plaintext json config
	if err := json.Unmarshal(cfgBytes, &resp.ChannelConfig); err != nil {
		return NotificationConfigResponse{}, err
	}

	// Check for sensitive value
	pw, ok := resp.ChannelConfig["password"].(string)
	if ok && pw != "" {
		delete(resp.ChannelConfig, "password")
	}

	return resp, nil
}

// Add or update personal notification options
func (s *Service) setUserNotificationSettings(userID int64, req NotificationConfigRequest) error {
	// Check the notification type is supported
	handler, ok := registry[req.Type]
	if !ok {
		return shared.ErrNotificationTypeNotSupported
	}

	// Call handler specific validation
	if err := handler.Validate(req.ChannelConfig); err != nil {
		return err
	}

	// Call handler specific marshaler
	cfgStr, err := handler.ToConfig(req.ChannelConfig)
	if err != nil {
		return err
	}

	// Encrypt notification config
	encryptedBytes, err := s.Encryptor.Encrypt([]byte(cfgStr))
	if err != nil {
		return err
	}

	settings := NotificationSetting{
		UserID:    userID,
		Type:      req.Type,
		EncConfig: encryptedBytes,
	}

	// Store it
	if err := s.Repo.setUserNotificationSettings(settings); err != nil {
		return err
	}

	return nil
}

func (s *Service) deleteUserNotificationSettings(userID int64) error {
	if err := s.Repo.deleteUserNotificationSettings(userID); err != nil {
		return err
	}

	return nil
}

// Decrypt all existing notification configs and re-encrypt using supplied phrase
func (s *Service) rotateEncryptionPhrase(req RotateEncryptionPhraseRequest) error {
	if req.Phrase == "" {
		return shared.ErrFieldMissing
	}
	// Get settings encrypted with old key
	settings, err := s.Repo.getAllUserNotificationSettings()
	if err != nil {
		return err
	}

	// Create new encryptor
	encryptor, err := encryption.NewAESGCMEncryptor(req.Phrase)
	if err != nil {
		return err
	}

	// all or nothing
	tx, err := s.Repo.Base.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, setting := range settings {
		// Use app decryptor to decrypt
		cfgBytes, err := s.Encryptor.Decrypt(setting.EncConfig)
		if err != nil {
			return err
		}

		// Encrypt using new encryptor
		encryptedBytes, err := encryptor.Encrypt(cfgBytes)
		if err != nil {
			return err
		}

		setting.EncConfig = encryptedBytes

		// Write
		if err := s.Repo.setUserNotificationSettingsTx(tx, setting); err != nil {
			return err
		}
	}

	tx.Commit()

	// User now replaces env var with new phrase and restarts app
	return nil
}
