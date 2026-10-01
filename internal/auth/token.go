package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateToken generates a random 32-byte token and its SHA-256 hash.
func GenerateToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Create a user-friendly token format: hl_<base64>
	tokenValue := base64.RawURLEncoding.EncodeToString(bytes)
	token := "hl_" + tokenValue
	
	hash := HashToken(token)
	return token, hash, nil
}

// HashToken creates a SHA-256 hash of the plain token.
func HashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}

// CheckTokenHash compares a plain token against a hash in constant time.
func CheckTokenHash(plainToken, expectedHash string) bool {
	actualHash := HashToken(plainToken)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}
