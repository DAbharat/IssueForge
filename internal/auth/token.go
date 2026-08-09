package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateRefToken() (string, error) {
	r := make([]byte, 32)

	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	token := hex.EncodeToString(r)

	return token, nil
}
