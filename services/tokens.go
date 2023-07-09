package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateWSToken(length int) (string, error) {
	// Calculate the byte size needed for the token
	byteSize := (length * 3) / 4

	// Create a byte slice with the required size
	tokenBytes := make([]byte, byteSize)

	// Generate random bytes
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes as a base64 string
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token[:length], nil
}

func GenerateApiToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	token := base64.URLEncoding.EncodeToString(hash[:])

	return token, err
}
