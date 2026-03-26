package encryption

import (
	"crypto/cipher"
	"ez2boot/internal/audit"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/auth/oidc"
	"ez2boot/internal/db"
	"ez2boot/internal/notification"
	"log/slog"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo                *Repository
	NotificationService *notification.Service
	LdapService         *ldap.Service
	OidcService         *oidc.Service
	Audit               *audit.Service
	Encryptor           Encryptor
	Logger              *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type AESGCMEncryptor struct {
	gcm cipher.AEAD
}

type RotateEncryptionPhraseRequest struct {
	Phrase string `json:"phrase"`
}
