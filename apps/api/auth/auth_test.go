//nolint:testpackage,exhaustruct
package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	gojwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/jackc/pgx/v5"

	dbqueries "github.com/DomNidy/saint_sim/internal/db"
)

type stubAPIKeyAuthenticator struct {
	authenticate func(context.Context, string) (AuthContext, error)
}

func (stub stubAPIKeyAuthenticator) Authenticate(
	ctx context.Context,
	rawAPIKey string,
) (AuthContext, error) {
	return stub.authenticate(ctx, rawAPIKey)
}

type stubJWTAuthenticator struct {
	verify func(context.Context, string) (AuthContext, error)
}

func (stub stubJWTAuthenticator) Authenticate(
	ctx context.Context,
	rawToken string,
) (AuthContext, error) {
	return stub.verify(ctx, rawToken)
}

type stubJWTKeyLookup struct {
	get func(context.Context, string) (dbqueries.Jwk, error)
}

func (stub stubJWTKeyLookup) GetJwkByID(ctx context.Context, id string) (dbqueries.Jwk, error) {
	return stub.get(ctx, id)
}

type stubAPIKeyLookup struct {
	getAPIKey func(context.Context, string) (dbqueries.GetApiKeyRow, error)
}

//nolint:revive // method name must match the production interface
func (stub stubAPIKeyLookup) GetApiKey(
	ctx context.Context,
	apiKey string,
) (dbqueries.GetApiKeyRow, error) {
	return stub.getAPIKey(ctx, apiKey)
}

func TestJWTAuthenticatorRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	publicKeyJSON, signedToken := signedTestJWT(
		t,
		"test-kid",
		"https://auth.example.com",
		"wrong-audience",
		"user-42",
	)
	verifier := NewJWTAuthenticator(
		stubJWTKeyLookup{
			get: func(context.Context, string) (dbqueries.Jwk, error) {
				return dbqueries.Jwk{
					ID:        "test-kid",
					PublicKey: publicKeyJSON,
				}, nil
			},
		},
		&gojwt.Expected{AnyAudience: []string{"saint-api"}, Issuer: "https://auth.example.com"},
	)

	_, err := verifier.Authenticate(t.Context(), signedToken)
	if !errors.Is(err, errInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token, got %v", err)
	}
}

func TestJWTAuthenticatorRejectsUnknownKeyID(t *testing.T) {
	t.Parallel()

	_, signedToken := signedTestJWT(
		t,
		"missing-kid",
		"https://auth.example.com",
		"saint-api",
		"user-42",
	)
	verifier := NewJWTAuthenticator(
		stubJWTKeyLookup{
			get: func(context.Context, string) (dbqueries.Jwk, error) {
				return dbqueries.Jwk{}, pgx.ErrNoRows
			},
		},
		&gojwt.Expected{AnyAudience: []string{"saint-api"}, Issuer: "https://auth.example.com"},
	)

	_, err := verifier.Authenticate(t.Context(), signedToken)
	if !errors.Is(err, errInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token, got %v", err)
	}
}

func TestSliceSecretFromApiKey(t *testing.T) {
	t.Parallel()
	type expectedResult struct {
		input  string
		secret string
		err    error
	}

	// #nosec G101 -- test fixture
	cases := []expectedResult{
		{
			input:  "sk_live_ae2313f129305104310",
			secret: "ae2313f129305104310",
			err:    nil,
		},
		{
			input:  "sk_org_live_test_12345abc",
			secret: "12345abc",
			err:    nil,
		},
		{
			input:  "sk_test_",
			secret: "",
			err:    errMalformedAPIKey,
		},
		{
			input:  "171c2edc4bf65b068537e593d0650f86ead7fa1f0cd4255e8d2cec1022a32cd9",
			secret: "",
			err:    errMalformedAPIKey,
		},
	}

	for _, testCase := range cases {
		secret, err := sliceSecretFromAPIKey(testCase.input)
		if secret != testCase.secret {
			t.Fatalf(
				"Extracted secret '%s' did not match expected '%s'. Input: '%s'",
				secret,
				testCase.secret,
				testCase.input,
			)
		}
		if !errors.Is(err, testCase.err) {
			t.Fatalf(
				"Expected error %v, but got %v. Input: '%v'",
				testCase.err,
				err,
				testCase.input,
			)
		}
	}
}

func TestAPIKeyAuthenticatorRejectsMalformedAPIKey(t *testing.T) {
	t.Parallel()

	authenticator := NewAPIKeyAuthenticator(stubAPIKeyLookup{
		getAPIKey: func(context.Context, string) (dbqueries.GetApiKeyRow, error) {
			t.Fatal("GetApiKey should not be called for malformed keys")

			return dbqueries.GetApiKeyRow{}, nil
		},
	})

	_, err := authenticator.Authenticate(
		t.Context(),
		"171c2edc4bf65b068537e593d0650f86ead7fa1f0cd4255e8d2cec1022a32cd9",
	)
	if !errors.Is(err, errMalformedAPIKey) {
		t.Fatalf("expected malformed api key, got %v", err)
	}
}

func TestAPIKeyAuthenticatorRejectsUnknownWellFormedAPIKey(t *testing.T) {
	t.Parallel()

	authenticator := NewAPIKeyAuthenticator(stubAPIKeyLookup{
		getAPIKey: func(context.Context, string) (dbqueries.GetApiKeyRow, error) {
			return dbqueries.GetApiKeyRow{}, pgx.ErrNoRows
		},
	})

	_, err := authenticator.Authenticate(t.Context(), "sk_test_12345abc")
	if !errors.Is(err, errInvalidAPIKey) {
		t.Fatalf("expected invalid api key, got %v", err)
	}
}

func TestEffectiveUserID(t *testing.T) {
	t.Parallel()

	userID := "user-42"

	cases := []struct {
		name        string
		authContext AuthContext
		expectedID  string
		expectedOK  bool
	}{
		{
			name: "bearer auth",
			authContext: AuthContext{
				Scheme: AuthSchemeBearer,
				UserID: userID,
			},
			expectedID: userID,
			expectedOK: true,
		},
		{
			name: "user-owned api key",
			authContext: AuthContext{
				Scheme: AuthSchemeAPIKey,
				APIKey: &dbqueries.GetApiKeyRow{
					PrincipalType: dbqueries.PrincipalTypeUser,
					UserID:        &userID,
				},
			},
			expectedID: userID,
			expectedOK: true,
		},
		{
			name: "service-owned api key",
			authContext: AuthContext{
				Scheme: AuthSchemeAPIKey,
				APIKey: &dbqueries.GetApiKeyRow{
					PrincipalType: dbqueries.PrincipalTypeService,
				},
			},
			expectedID: "",
			expectedOK: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualID, actualOK := EffectiveUserID(testCase.authContext)
			if actualID != testCase.expectedID || actualOK != testCase.expectedOK {
				t.Fatalf(
					"EffectiveUserID() = (%q, %v), want (%q, %v)",
					actualID,
					actualOK,
					testCase.expectedID,
					testCase.expectedOK,
				)
			}
		})
	}
}

func signedTestJWT(
	t *testing.T,
	keyID string,
	issuer string,
	audience string,
	subject string,
) (string, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key pair: %v", err)
	}

	publicJWK := jose.JSONWebKey{
		Key:       publicKey,
		KeyID:     keyID,
		Algorithm: string(jose.EdDSA),
		Use:       "sig",
	}
	rawPublicJWK, err := json.Marshal(publicJWK)
	if err != nil {
		t.Fatalf("marshal public jwk: %v", err)
	}

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.EdDSA,
		Key: jose.JSONWebKey{
			Key:       privateKey,
			KeyID:     keyID,
			Algorithm: string(jose.EdDSA),
		},
	}, nil)
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}

	token, err := gojwt.Signed(signer).Claims(gojwt.Claims{
		Issuer:   issuer,
		Subject:  subject,
		Audience: gojwt.Audience{audience},
		IssuedAt: gojwt.NewNumericDate(time.Now()),
		Expiry:   gojwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	}).Serialize()
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return string(rawPublicJWK), token
}
