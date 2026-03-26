package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"ez2boot/internal/audit"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/auth/oidc"
	"ez2boot/internal/db"
	"ez2boot/internal/notification"
	"log/slog"
)

func NewHandler(encryptionService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: encryptionService,
		Logger:  logger,
	}
}

func NewService(encryptionRepo *Repository, notificationService *notification.Service, ldapService *ldap.Service, oidcService *oidc.Service, audit *audit.Service, encryptor Encryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:                encryptionRepo,
		NotificationService: notificationService,
		LdapService:         ldapService,
		OidcService:         oidcService,
		Audit:               audit,
		Encryptor:           encryptor,
		Logger:              logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}

// Create new encryptor from passphrase
func NewAESGCMEncryptor(passphrase string) (*AESGCMEncryptor, error) {
	key := sha256.Sum256([]byte(passphrase)) // Get 256 bits from passphrase
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESGCMEncryptor{gcm: gcm}, nil
}
