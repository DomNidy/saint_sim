package secrets

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load secrets into memory when this package is imported
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error occured while loading .env", err)
	}
}

type Secret interface {
	// Returns the key the secret (the same as the key in the .env file we load these from)
	Key() string
	// Returns the actual value of the secret
	Value() string
	// Returns the value of the secret, masking out all characters after the 3rd
	MaskedValue() string
}

type SecretImpl struct {
	key   string
	value string
}

func (s *SecretImpl) Key() string {
	return s.key
}

func (s *SecretImpl) Value() string {
	return s.value
}

func (s *SecretImpl) MaskedValue() string {
	return maskToken(s.value, 3)
}

func NewSecret(key, value string) Secret {
	return &SecretImpl{key: key, value: value}
}

// Panics on error
func LoadSecret(key string) Secret {
	var secret Secret = nil

	if value, exists := os.LookupEnv(key); exists {
		secret = NewSecret(key, value)
	} else {
		panic(fmt.Sprintf("failed to load secret with key: %s from env file", key))
	}
	return secret
}

// Used to print out the secrets to console
func maskToken(token string, visibleChars int) string {
	if len(token) < 3 {
		return token
	}
	return token[:visibleChars] + "XXX"
}
