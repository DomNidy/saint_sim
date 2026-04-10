package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
)

// APIKeyAuthenticator validates Api-Key credentials and resolves their owning principal.
type APIKeyAuthenticator struct {
	lookup APIKeyLookup
}

// NewAPIKeyAuthenticator builds an API key authenticator that resolves API keys using the
// provided APIKeyLookup.
func NewAPIKeyAuthenticator(lookup APIKeyLookup) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{lookup: lookup}
}

// Authenticate validates a raw API key and returns the resolved request auth context.
func (validator *APIKeyAuthenticator) Authenticate(
	ctx context.Context,
	rawAPIKey string,
) (AuthContext, error) {
	secretPortion, ok := sliceSecretFromAPIKey(rawAPIKey)
	if !ok {
		return AuthContext{}, errInvalidAPIKey
	}

	hashedAPIKey := api_utils.HashAPIKey(secretPortion)

	resAPIKey, err := validator.lookup.GetApiKey(ctx, hashedAPIKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthContext{}, errInvalidAPIKey
		}

		return AuthContext{}, fmt.Errorf("error occurred while looking up API key: %w", err)
	}

	// sanity check
	if resAPIKey.SecretHash != hashedAPIKey {
		return AuthContext{}, errAPIKeySanityCheckFail
	}

	return AuthContext{
		Scheme: AuthSchemeAPIKey,
		APIKey: &resAPIKey,
		UserID: "",
	}, nil
}

// sliceSecretFromAPIKey extracts the "secret" portion a raw API key string.
// example: "sk_test_12345" -> "12345" is returned.
func sliceSecretFromAPIKey(rawAPIKey string) (string, bool) {
	lastIndex := strings.LastIndex(rawAPIKey, "_")

	if lastIndex == -1 || lastIndex == len(rawAPIKey)-1 {
		return "", false
	}

	secretPart := rawAPIKey[lastIndex+1:]

	return secretPart, true
}
