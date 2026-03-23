package encryption

import (
	"context"
	"database/sql"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
)

func (s *Service) rotateEncryptionPhrase(req RotateEncryptionPhraseRequest, ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "rotate",
			Resource:    "encryption phrase",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	if req.Phrase == "" {
		return shared.ErrFieldMissing
	}

	// Create new encryptor
	newEncryptor, err := NewAESGCMEncryptor(req.Phrase)
	if err != nil {
		return err
	}

	// all or nothing
	tx, err := s.Repo.Base.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// User notification settings
	if err := s.reEncryptNotificationSettings(tx, newEncryptor); err != nil {
		return err
	}

	// Ldap password
	if err := s.reEncryptLdapPassword(tx, newEncryptor); err != nil {
		return err
	}

	// Oidc secret
	if err := s.reEncryptOidcSecret(tx, newEncryptor); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// User now replaces env var with new phrase and restarts app
	return nil
}

func (s *Service) reEncryptNotificationSettings(tx *sql.Tx, newEncryptor Encryptor) error {
	// Get settings encrypted with old key
	settings, err := s.NotificationService.GetAllUserNotificationSettings()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		// Use app decryptor to decrypt
		cfgBytes, err := s.Encryptor.Decrypt(setting.EncConfig)
		if err != nil {
			return err
		}

		// Encrypt using new encryptor
		encryptedBytes, err := newEncryptor.Encrypt(cfgBytes)
		if err != nil {
			return err
		}

		setting.EncConfig = encryptedBytes

		// Write
		if err := s.NotificationService.SetUserNotificationSettingsTx(tx, setting); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) reEncryptLdapPassword(tx *sql.Tx, newEncryptor Encryptor) error {
	// Get settings encrypted with old key
	encPassword, err := s.LdapService.GetLdapPassword()
	if err != nil {
		return err
	}

	// Use app decryptor to decrypt
	password, err := s.Encryptor.Decrypt(encPassword)
	if err != nil {
		return err
	}

	// Encrypt using new encryptor
	encryptedBytes, err := newEncryptor.Encrypt(password)
	if err != nil {
		return err
	}

	// Write
	if err := s.LdapService.SetLdapPasswordTx(tx, encryptedBytes); err != nil {
		return err
	}

	return nil
}

func (s *Service) reEncryptOidcSecret(tx *sql.Tx, newEncryptor Encryptor) error {
	// Get settings encrypted with old key
	encSecret, err := s.OidcService.GetOidcSecret()
	if err != nil {
		return err
	}

	// Use app decryptor to decrypt
	secret, err := s.Encryptor.Decrypt(encSecret)
	if err != nil {
		return err
	}

	// Encrypt using new encryptor
	encryptedBytes, err := newEncryptor.Encrypt(secret)
	if err != nil {
		return err
	}

	// Write
	if err := s.OidcService.SetOidcSecretTx(tx, encryptedBytes); err != nil {
		return err
	}

	return nil
}
