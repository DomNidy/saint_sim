package api_utils

import (
	"encoding/hex"

	"crypto/sha256"

	uuid "github.com/google/uuid"
)

// GenerateUUID generates a UUID and returns it as a string.
func GenerateUUID() string {
	return uuid.New().String()
}

// HashAPIKey returns a SHA256 hash of an API key string.
//
// You should hash only the secret portion of the string;
// the stored secret hash in the db excludes the prefix.
// i.e., remove the "sk_xxx_" prefix, then hash the
// resulting string.
func HashAPIKey(apiKey string) string {
	bytes := sha256.Sum256([]byte(apiKey))

	return hex.EncodeToString(bytes[:])
}
