// Package middleware provides middleware that is used to authenticate incoming API requests.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

// AuthScheme identifies which authentication scheme authorized the request.
type AuthScheme string

const (
	// AuthSchemeAPIKey identifies requests authenticated by the Api-Key header.
	AuthSchemeAPIKey AuthScheme = "api_key"
	// AuthSchemeBearer identifies requests authenticated by a bearer JWT.
	AuthSchemeBearer AuthScheme = "bearer"

	// Key used in Gin request Context to store the principal used to
	// authenticate a request. Principal is a normalized object that
	// we create after successful authentication.
	//
	// Using a normalized principal like this is helpful as we support
	// authentication with multiple schemes (api_key, bearer JWT), so
	// subsequent server code doesn't need to re-lower the scheme.
	authPrincipalContextKey = "auth.principal"
)

var (
	errAPIKeySanityCheckFail       = errors.New("api key sanity check failed")
	errInvalidAPIKey               = errors.New("invalid api key")
	errInvalidBearerToken          = errors.New("invalid bearer token")
	errMalformedAuthorizationValue = errors.New("malformed authorization header")
	errMissingCredentials          = errors.New("missing credentials")
)

// AuthPrincipal stores the resolved authentication identity for a request.
type AuthPrincipal struct {
	Scheme AuthScheme

	// If the scheme was bearer, the user ID that was encoded
	// in the JWT. Otherwise, this is empty.
	UserID string
}

// APIKeyLookup loads API keys from the backing store.
type APIKeyLookup interface {
	GetApiKey(ctx context.Context, apiKey string) (dbqueries.ApiKey, error)
}

// APIKeyValidator validates raw API keys from incoming requests.
type APIKeyValidator interface {
	Validate(ctx context.Context, rawAPIKey string) error
}

// JWTVerifier validates bearer JWTs and returns the authenticated principal.
type JWTVerifier interface {
	Verify(ctx context.Context, rawToken string) (AuthPrincipal, error)
}

type dbAPIKeyValidator struct {
	lookup APIKeyLookup
}

// NewAPIKeyValidator builds an API key validator backed by the database.
func NewAPIKeyValidator(lookup APIKeyLookup) APIKeyValidator {
	return &dbAPIKeyValidator{lookup: lookup}
}

// AuthRequire validates that incoming requests provide either a valid Api-Key or a valid Bearer
// JWT.
func AuthRequire(apiKeyValidator APIKeyValidator, jwtVerifier JWTVerifier) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		apiKey := strings.TrimSpace(ginContext.GetHeader("Api-Key"))
		authorizationValue := strings.TrimSpace(ginContext.GetHeader("Authorization"))

		principal, err := authenticateRequest(
			ginContext.Request.Context(),
			apiKeyValidator,
			jwtVerifier,
			apiKey,
			authorizationValue,
		)
		if err == nil {
			ginContext.Set(authPrincipalContextKey, principal)
			ginContext.Next()

			return
		}

		if errors.Is(err, errInvalidAPIKey) ||
			errors.Is(err, errInvalidBearerToken) ||
			errors.Is(
				err,
				errMalformedAuthorizationValue,
			) || errors.Is(err, errMissingCredentials) {
			log.Printf("Unauthorized request: %v", err)
			abortWithError(ginContext, http.StatusUnauthorized, "Unauthorized")

			return
		}

		log.Printf("Internal auth error: %v", err)
		abortWithError(ginContext, http.StatusInternalServerError, "Internal server error")
	}
}

// GetAuthPrincipal retrieves the authenticated request principal from Gin context.
func GetAuthPrincipal(ginContext *gin.Context) (AuthPrincipal, bool) {
	rawPrincipal, exists := ginContext.Get(authPrincipalContextKey)
	if !exists {
		return AuthPrincipal{Scheme: "", UserID: ""}, false
	}

	principal, ok := rawPrincipal.(AuthPrincipal)
	if !ok {
		return AuthPrincipal{Scheme: "", UserID: ""}, false
	}

	return principal, true
}

func (validator *dbAPIKeyValidator) Validate(ctx context.Context, rawAPIKey string) error {
	secretPortion, ok := sliceSecretFromAPIKey(rawAPIKey)
	if !ok {
		return fmt.Errorf("invalid API key")
	}

	hashedAPIKey := api_utils.HashAPIKey(secretPortion)

	resAPIKey, err := validator.lookup.GetApiKey(ctx, hashedAPIKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidAPIKey
		}

		return fmt.Errorf("error occurred while looking up API key: %w", err)
	}

	// sanity check
	if resAPIKey.SecretHash != hashedAPIKey {
		return errAPIKeySanityCheckFail
	}

	return nil
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

// authenticateRequest authenticates an incoming request. Supports both API key and
// bearer JWT authentication.
//
// The returned AuthPrincipal provides context on what scheme was used to authenticate the
// request (and additional context, such as the user id in the case of bearer JWT auth).
//
//nolint:cyclop // This function intentionally implements fallback auth scheme evaluation.
func authenticateRequest(
	ctx context.Context,
	apiKeyValidator APIKeyValidator,
	jwtVerifier JWTVerifier,
	apiKey string,
	authorizationValue string,
) (AuthPrincipal, error) {
	var internalErrors []error

	var invalidCredentialErrors []error

	if authorizationValue != "" {
		rawToken, err := bearerTokenFromHeader(authorizationValue)
		switch {
		case err == nil:
			principal, verifyErr := jwtVerifier.Verify(ctx, rawToken)
			switch {
			case verifyErr == nil:
				return principal, nil
			case errors.Is(verifyErr, errInvalidBearerToken):
				invalidCredentialErrors = append(invalidCredentialErrors, verifyErr)
			default:
				internalErrors = append(internalErrors, verifyErr)
			}
		case errors.Is(err, errMalformedAuthorizationValue):
			invalidCredentialErrors = append(invalidCredentialErrors, err)
		default:
			internalErrors = append(internalErrors, err)
		}
	}

	if apiKey != "" {
		err := apiKeyValidator.Validate(ctx, apiKey)
		switch {
		case err == nil:
			return AuthPrincipal{Scheme: AuthSchemeAPIKey, UserID: ""}, nil
		case errors.Is(err, errInvalidAPIKey):
			invalidCredentialErrors = append(invalidCredentialErrors, err)
		default:
			internalErrors = append(internalErrors, err)
		}
	}

	if len(internalErrors) > 0 {
		return AuthPrincipal{}, errors.Join(internalErrors...)
	}

	if len(invalidCredentialErrors) > 0 {
		return AuthPrincipal{}, errors.Join(invalidCredentialErrors...)
	}

	return AuthPrincipal{Scheme: "", UserID: ""}, errMissingCredentials
}

// bearerTokenFromHeader parses the <authToken> from the `Bearer <authToken>` header string.
// pass this function the value of the `Authorization` header.
func bearerTokenFromHeader(authorizationValue string) (string, error) {
	scheme, token, found := strings.Cut(authorizationValue, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", errMalformedAuthorizationValue
	}

	return strings.TrimSpace(token), nil
}

func abortWithError(ginContext *gin.Context, statusCode int, message string) {
	ginContext.AbortWithStatusJSON(statusCode, api_types.ErrorResponse{
		Message: utils.StrPtr(message),
	})
}
