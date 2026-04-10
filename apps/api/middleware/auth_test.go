//nolint:testpackage,exhaustruct,varnamelen,noctx,wsl_v5
package middleware

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	gojwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
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

type stubJWTVerifier struct {
	verify func(context.Context, string) (AuthContext, error)
}

func (stub stubJWTVerifier) Verify(ctx context.Context, rawToken string) (AuthContext, error) {
	return stub.verify(ctx, rawToken)
}

type stubJWTKeyLookup struct {
	get func(context.Context, string) (dbqueries.Jwk, error)
}

func (stub stubJWTKeyLookup) GetJwkByID(ctx context.Context, id string) (dbqueries.Jwk, error) {
	return stub.get(ctx, id)
}

func TestAuthRequireAcceptsBearerJWT(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequire(
		stubAPIKeyAuthenticator{authenticate: func(context.Context, string) (AuthContext, error) {
			t.Fatal("api key validator should not run when bearer auth succeeds")

			return AuthContext{}, nil
		}},
		stubJWTVerifier{verify: func(_ context.Context, rawToken string) (AuthContext, error) {
			if rawToken != "valid-token" {
				t.Fatalf("unexpected token %q", rawToken)
			}

			return AuthContext{
				Scheme: AuthSchemeBearer,
				UserID: "user_123",
			}, nil
		}},
	))
	router.POST("/", func(c *gin.Context) {
		authContext, ok := GetAuthContext(c)
		if !ok {
			t.Fatal("auth context missing")
		}
		if authContext.UserID != "user_123" {
			t.Fatalf("unexpected user id %q", authContext.UserID)
		}

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
}

func TestAuthRequireFallsBackToAPIKey(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequire(
		stubAPIKeyAuthenticator{authenticate: func(_ context.Context, rawAPIKey string) (AuthContext, error) {
			if rawAPIKey != "good-api-key" {
				t.Fatalf("unexpected api key %q", rawAPIKey)
			}

			userID := "user_123"

			return AuthContext{
				Scheme: AuthSchemeAPIKey,
				APIKey: &dbqueries.GetApiKeyRow{
					SecretHash:    "hashed-key",
					PrincipalID:   uuid.MustParse("3ef10f0f-cd18-454d-8724-e0d9d3ac67bf"),
					PrincipalType: dbqueries.PrincipalTypeUser,
					UserID:        &userID,
					ServiceID:     nil,
				},
			}, nil
		}},
		stubJWTVerifier{verify: func(context.Context, string) (AuthContext, error) {
			return AuthContext{Scheme: "", UserID: "", APIKey: nil}, errInvalidBearerToken
		}},
	))
	router.POST("/", func(c *gin.Context) {
		authContext, ok := GetAuthContext(c)
		if !ok {
			t.Fatal("auth context missing")
		}
		if authContext.Scheme != AuthSchemeAPIKey {
			t.Fatalf("unexpected auth scheme %q", authContext.Scheme)
		}
		if authContext.APIKey == nil || authContext.APIKey.PrincipalType != dbqueries.PrincipalTypeUser {
			t.Fatalf("expected user-owned api key auth context, got %#v", authContext.APIKey)
		}

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	req.Header.Set("Api-Key", "good-api-key")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
}

func TestAuthRequireRejectsMissingCredentials(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequire(
		stubAPIKeyAuthenticator{authenticate: func(context.Context, string) (AuthContext, error) {
			return AuthContext{}, nil
		}},
		stubJWTVerifier{verify: func(context.Context, string) (AuthContext, error) {
			return AuthContext{Scheme: "", UserID: "", APIKey: nil}, nil
		}},
	))
	router.POST("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
}

func TestJWTVerifierValidatesClaimsAndSignature(t *testing.T) {
	t.Parallel()

	publicKeyJSON, signedToken := signedTestJWT(
		t,
		"test-kid",
		"https://auth.example.com",
		"saint-api",
		"user-42",
	)
	verifier := NewJWTVerifier(
		stubJWTKeyLookup{
			get: func(_ context.Context, id string) (dbqueries.Jwk, error) {
				if id != "test-kid" {
					t.Fatalf("unexpected key id %q", id)
				}

				return dbqueries.Jwk{
					ID:        "test-kid",
					PublicKey: publicKeyJSON,
				}, nil
			},
		},
		"https://auth.example.com",
		"saint-api",
	)

	principal, err := verifier.Verify(context.Background(), signedToken)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if principal.Scheme != AuthSchemeBearer {
		t.Fatalf("unexpected scheme %q", principal.Scheme)
	}
	if principal.UserID != "user-42" {
		t.Fatalf("unexpected user id %q", principal.UserID)
	}
}

func TestAuthRequireAcceptsServiceOwnedAPIKey(t *testing.T) {
	t.Parallel()

	serviceID := uuid.MustParse("0d8be06d-a5d2-45ed-b462-1b310813210f")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequire(
		stubAPIKeyAuthenticator{authenticate: func(_ context.Context, rawAPIKey string) (AuthContext, error) {
			if rawAPIKey != "service-api-key" {
				t.Fatalf("unexpected api key %q", rawAPIKey)
			}

			return AuthContext{
				Scheme: AuthSchemeAPIKey,
				APIKey: &dbqueries.GetApiKeyRow{
					SecretHash:    "hashed-key",
					PrincipalID:   uuid.MustParse("4d3c35af-430f-45de-9d12-ec3b84db5042"),
					PrincipalType: dbqueries.PrincipalTypeService,
					UserID:        nil,
					ServiceID:     &serviceID,
				},
			}, nil
		}},
		stubJWTVerifier{verify: func(context.Context, string) (AuthContext, error) {
			return AuthContext{Scheme: "", UserID: "", APIKey: nil}, errInvalidBearerToken
		}},
	))
	router.POST("/", func(c *gin.Context) {
		authContext, ok := GetAuthContext(c)
		if !ok {
			t.Fatal("auth context missing")
		}
		if authContext.APIKey == nil || authContext.APIKey.PrincipalType != dbqueries.PrincipalTypeService {
			t.Fatalf("expected service-owned api key auth context, got %#v", authContext.APIKey)
		}

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Api-Key", "service-api-key")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
}

func TestJWTVerifierRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	publicKeyJSON, signedToken := signedTestJWT(
		t,
		"test-kid",
		"https://auth.example.com",
		"wrong-audience",
		"user-42",
	)
	verifier := NewJWTVerifier(
		stubJWTKeyLookup{
			get: func(context.Context, string) (dbqueries.Jwk, error) {
				return dbqueries.Jwk{
					ID:        "test-kid",
					PublicKey: publicKeyJSON,
				}, nil
			},
		},
		"https://auth.example.com",
		"saint-api",
	)

	_, err := verifier.Verify(context.Background(), signedToken)
	if !errors.Is(err, errInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token, got %v", err)
	}
}

func TestJWTVerifierRejectsUnknownKeyID(t *testing.T) {
	t.Parallel()

	_, signedToken := signedTestJWT(
		t,
		"missing-kid",
		"https://auth.example.com",
		"saint-api",
		"user-42",
	)
	verifier := NewJWTVerifier(
		stubJWTKeyLookup{
			get: func(context.Context, string) (dbqueries.Jwk, error) {
				return dbqueries.Jwk{}, pgx.ErrNoRows
			},
		},
		"https://auth.example.com",
		"saint-api",
	)

	_, err := verifier.Verify(context.Background(), signedToken)
	if !errors.Is(err, errInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token, got %v", err)
	}
}

func TestSliceSecretFromApiKey(t *testing.T) {
	t.Parallel()
	type expectedResult struct {
		input  string
		secret string
		ok     bool
	}

	// #nosec G101 -- test fixture
	cases := []expectedResult{
		{
			input:  "sk_live_ae2313f129305104310",
			secret: "ae2313f129305104310",
			ok:     true,
		},
		{
			input:  "sk_org_live_test_12345abc",
			secret: "12345abc",
			ok:     true,
		},
		{
			input:  "sk_test_",
			secret: "",
			ok:     false,
		},
	}

	for _, testCase := range cases {
		secret, ok := sliceSecretFromAPIKey(testCase.input)
		if secret != testCase.secret {
			t.Fatalf(
				"Extracted secret '%s' did not match expected '%s'. Input: '%s'",
				secret,
				testCase.secret,
				testCase.input,
			)
		}
		if ok != testCase.ok {
			t.Fatalf(
				"Expected %v, but got %v. Input: '%v'",
				testCase.ok,
				ok,
				testCase.input,
			)
		}
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
		testCase := testCase

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
