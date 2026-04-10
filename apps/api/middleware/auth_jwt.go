package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	gojwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/jackc/pgx/v5"

	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
)

var errEmptyPublicKey = errors.New("public key is empty")

// JWTKeyLookup loads signing keys from the backing store by key ID.
type JWTKeyLookup interface {
	GetJwkByID(ctx context.Context, id string) (dbqueries.Jwk, error)
}

type dbJWTVerifier struct {
	keyLookup JWTKeyLookup
	issuer    string
	audience  string
}

// NewJWTVerifier builds a JWT verifier backed by keys from the jwks table.
func NewJWTVerifier(keyLookup JWTKeyLookup, issuer string, audience string) JWTVerifier {
	return &dbJWTVerifier{
		keyLookup: keyLookup,
		issuer:    issuer,
		audience:  audience,
	}
}

//nolint:cyclop // JWT verification intentionally checks several independent failure modes.
func (verifier *dbJWTVerifier) Verify(ctx context.Context, rawToken string) (AuthContext, error) {
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

	jwkRecord, err := verifier.keyLookup.GetJwkByID(ctx, keyID)
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

	expectedClaims := gojwt.Expected{
		Issuer:      verifier.issuer,
		Subject:     "",
		AnyAudience: nil,
		ID:          "",
		Time:        time.Now(),
	}

	err = claims.ValidateWithLeeway(expectedClaims, 0)
	if err != nil {
		return AuthContext{}, fmt.Errorf(
			"%w: validate standard claims: %w",
			errInvalidBearerToken,
			err,
		)
	}

	if !claims.Audience.Contains(verifier.audience) {
		return AuthContext{}, fmt.Errorf("%w: unexpected audience", errInvalidBearerToken)
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return AuthContext{}, fmt.Errorf("%w: missing subject", errInvalidBearerToken)
	}

	return AuthContext{
		Scheme: AuthSchemeBearer,
		UserID: claims.Subject,
	}, nil
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
