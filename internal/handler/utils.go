package handler

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateToken(n int) (string, error) {
	randomBytes := make([]byte, n)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
