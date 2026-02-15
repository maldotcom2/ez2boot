package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
)

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
