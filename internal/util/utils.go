package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/alexedwards/argon2id"
)

func GenerateRandomString(n int) (string, error) {
	randomBytes := make([]byte, n)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// Password hashing only
func HashPassword(secret string) (string, error) {
	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(secret, params)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// Deterministic hashing for session tokens
func HashToken(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	tokenHash := hex.EncodeToString(hash[:])

	return tokenHash
}

func GetExpiryFromDuration(currentExpiry int64, duration string) (int64, error) {

	dur, err := time.ParseDuration(duration) // Parse string (eg, 4h) to time type
	if err != nil {
		return 0, err
	}

	// Case for new session
	if currentExpiry == 0 {
		now := time.Now().UTC()
		newExpiry := now.Add(dur).Unix()

		return newExpiry, nil

		// Case for update session
	} else {
		exp := time.Unix(currentExpiry, 0).UTC()
		newExpiry := exp.Add(dur).Unix()

		return newExpiry, nil
	}
}
