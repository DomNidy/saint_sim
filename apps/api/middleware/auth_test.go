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
	"github.com/jackc/pgx/v5"

	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
)

type stubAPIKeyValidator struct {
	validate func(context.Context, string) error
}

func (stub stubAPIKeyValidator) Validate(ctx context.Context, rawAPIKey string) error {
	return stub.validate(ctx, rawAPIKey)
}

type stubJWTVerifier struct {
	verify func(context.Context, string) (AuthPrincipal, error)
}

func (stub stubJWTVerifier) Verify(ctx context.Context, rawToken string) (AuthPrincipal, error) {
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
		stubAPIKeyValidator{validate: func(context.Context, string) error {
			t.Fatal("api key validator should not run when bearer auth succeeds")

			return nil
		}},
		stubJWTVerifier{verify: func(_ context.Context, rawToken string) (AuthPrincipal, error) {
			if rawToken != "valid-token" {
				t.Fatalf("unexpected token %q", rawToken)
			}

			return AuthPrincipal{
				Scheme: AuthSchemeBearer,
				UserID: "user_123",
			}, nil
		}},
	))
	router.POST("/", func(c *gin.Context) {
		principal, ok := GetAuthPrincipal(c)
		if !ok {
			t.Fatal("auth principal missing")
		}
		if principal.UserID != "user_123" {
			t.Fatalf("unexpected user id %q", principal.UserID)
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
		stubAPIKeyValidator{validate: func(_ context.Context, rawAPIKey string) error {
			if rawAPIKey != "good-api-key" {
				t.Fatalf("unexpected api key %q", rawAPIKey)
			}

			return nil
		}},
		stubJWTVerifier{verify: func(context.Context, string) (AuthPrincipal, error) {
			return AuthPrincipal{Scheme: "", UserID: ""}, errInvalidBearerToken
		}},
	))
	router.POST("/", func(c *gin.Context) {
		principal, ok := GetAuthPrincipal(c)
		if !ok {
			t.Fatal("auth principal missing")
		}
		if principal.Scheme != AuthSchemeAPIKey {
			t.Fatalf("unexpected auth scheme %q", principal.Scheme)
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
		stubAPIKeyValidator{validate: func(context.Context, string) error {
			return nil
		}},
		stubJWTVerifier{verify: func(context.Context, string) (AuthPrincipal, error) {
			return AuthPrincipal{Scheme: "", UserID: ""}, nil
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

	publicKeyJSON, signedToken := signedTestJWT(t, "test-kid", "https://auth.example.com", "saint-api", "user-42")
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

func TestJWTVerifierRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	publicKeyJSON, signedToken := signedTestJWT(t, "test-kid", "https://auth.example.com", "wrong-audience", "user-42")
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

	_, signedToken := signedTestJWT(t, "missing-kid", "https://auth.example.com", "saint-api", "user-42")
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
