// Package secrets provides utilities for loading secrets into memory
package secrets

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultVisibleSecretChars = 3
	minMaskableTokenLength    = 3
)

// Load secrets into memory when this package is imported.
//
//nolint:gochecknoinits // makes sense to use init here imo
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("WARNING: failed to load .env. Secrets may not be populated: %v", err)
	}
}

// Secret is a key value pair, where the value is sensitive.
type Secret interface {
	// Returns the key the secret (the same as the key in the .env file we load these from)
	Key() string
	// Returns the actual value of the secret
	Value() string
	// Returns the value of the secret, masking out all characters after the 3rd
	MaskedValue() string
}

type secretImpl struct {
	key   string
	value string
}

func (s *secretImpl) Key() string {
	return s.key
}

func (s *secretImpl) Value() string {
	return s.value
}

func (s *secretImpl) MaskedValue() string {
	return MaskToken(s.value, defaultVisibleSecretChars)
}

func newSecret(key, value string) *secretImpl {
	return &secretImpl{key: key, value: value}
}

// LoadSecret returns the secret stored under key.
//
//nolint:ireturn // package callers intentionally depend on the Secret interface.
func LoadSecret(key string) Secret {
	if value, exists := os.LookupEnv(key); exists {
		return newSecret(key, value)
	}

	panic("failed to load secret with key: " + key)
}

// MaskToken returns token with only visibleChars characters preserved.
func MaskToken(token string, visibleChars int) string {
	if len(token) < minMaskableTokenLength {
		return "XXXXXXXX"
	}

	return token[:visibleChars] + "XXXXX"
}
