package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose/v4"
	gojwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/jackc/pgx/v5"

	dbqueries "github.com/DomNidy/saint_sim/pkg/db"
)

var errEmptyPublicKey = errors.New("public key is empty")

// JWTKeyLookup loads signing keys from the backing store by key ID.
type JWTKeyLookup interface {
	GetJwkByID(ctx context.Context, keyID string) (dbqueries.Jwk, error)
}

// JWTAuthenticator validates bearer JWTs and returns the authenticated context.
// Authenticate method returned an error if the JWT is invalid, or lacks the expected claims.
type JWTAuthenticator struct {
	keyLookup JWTKeyLookup
	claims    *gojwt.Expected
}

// NewJWTAuthenticator builds a JWT authenticator that enforces the provided
// claims and uses the key lookup to retrieve the signing key from the
// backing store by key ID.
//
// If no expected claims are provided, then the authenticator will pass as
// long as the key decodes.
func NewJWTAuthenticator(keyLookup JWTKeyLookup, claims *gojwt.Expected) *JWTAuthenticator {
	return &JWTAuthenticator{
		keyLookup: keyLookup,
		claims:    claims,
	}
}

// Authenticate verifies that the provided JWT token is valid and
// satisfies the claims expected by the JWTAuthenticator.
func (authenticator *JWTAuthenticator) Authenticate(
	ctx context.Context,
	rawToken string,
) (AuthContext, error) {
	parsedToken, err := gojwt.ParseSigned(rawToken, supportedJWTAlgorithms())
	if err != nil {
		return AuthContext{}, fmt.Errorf("%w: parse signed token: %w", errInvalidBearerToken, err)
	}

	if len(parsedToken.Headers) == 0 {
		return AuthContext{}, fmt.Errorf(
			"%w: token missing protected headers",
			errInvalidBearerToken,
		)
	}

	// we are using JWKs to validate token. This requires our JWT to have the
	// key ID header set.
	keyID := strings.TrimSpace(parsedToken.Headers[0].KeyID)
	if keyID == "" {
		return AuthContext{}, fmt.Errorf("%w: token missing key id", errInvalidBearerToken)
	}

	jwkRecord, err := authenticator.keyLookup.GetJwkByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthContext{}, fmt.Errorf("%w: unknown signing key", errInvalidBearerToken)
		}

		return AuthContext{}, fmt.Errorf("load jwk %q: %w", keyID, err)
	}

	publicKey, err := parseStoredPublicJWK(jwkRecord.PublicKey)
	if err != nil {
		return AuthContext{}, fmt.Errorf("parse jwk %q: %w", keyID, err)
	}

	var claims gojwt.Claims

	err = parsedToken.Claims(publicKey.Key, &claims)
	if err != nil {
		return AuthContext{}, fmt.Errorf(
			"%w: verify token signature: %w",
			errInvalidBearerToken,
			err,
		)
	}

	err = validateExpectedJWTClaims(claims, authenticator.claims)
	if err != nil {
		return AuthContext{}, err
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return AuthContext{}, fmt.Errorf("%w: missing subject", errInvalidBearerToken)
	}

	return AuthContext{
		Scheme: AuthSchemeBearer,
		UserID: claims.Subject,
		APIKey: nil,
	}, nil
}

func validateExpectedJWTClaims(claims gojwt.Claims, expected *gojwt.Expected) error {
	if expected == nil {
		return nil
	}

	err := claims.ValidateWithLeeway(*expected, 0)
	if err != nil {
		return fmt.Errorf("%w: validate standard claims: %w", errInvalidBearerToken, err)
	}

	return nil
}

func parseStoredPublicJWK(rawKey string) (jose.JSONWebKey, error) {
	var publicKey jose.JSONWebKey

	err := json.Unmarshal([]byte(rawKey), &publicKey)
	if err != nil {
		return jose.JSONWebKey{}, fmt.Errorf("unmarshal stored public jwk: %w", err)
	}

	if publicKey.Key == nil {
		return jose.JSONWebKey{}, errEmptyPublicKey
	}

	return publicKey, nil
}

func supportedJWTAlgorithms() []jose.SignatureAlgorithm {
	return []jose.SignatureAlgorithm{
		jose.EdDSA,
		jose.ES256,
		jose.ES512,
		jose.PS256,
		jose.RS256,
	}
}
