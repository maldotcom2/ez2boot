package encryption

import "crypto/cipher"

type AESGCMEncryptor struct {
	gcm cipher.AEAD
}
