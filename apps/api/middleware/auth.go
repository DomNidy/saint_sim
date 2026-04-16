// Package middleware provides middleware that is used to authenticate incoming API requests.
package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/internal/api_types"
	dbqueries "github.com/DomNidy/saint_sim/internal/db"
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
	authContextContextKey = "auth.context"
)

var (
	errAPIKeySanityCheckFail       = errors.New("api key sanity check failed")
	errInvalidAPIKey               = errors.New("invalid api key")
	errInvalidBearerToken          = errors.New("invalid bearer token")
	errMalformedAuthorizationValue = errors.New("malformed authorization header")
	errMissingCredentials          = errors.New("missing credentials")
)

// AuthContext stores the resolved authentication identity for a request.
type AuthContext struct {
	Scheme AuthScheme

	// If the scheme was bearer, the user ID that was encoded
	// in the JWT. Otherwise, this is empty.
	UserID string

	// APIKey contains the resolved API key owner when Api-Key auth succeeds.
	APIKey *dbqueries.GetApiKeyRow
}

// APIKeyLookup loads API keys and their owners from the backing store.
type APIKeyLookup interface {
	GetApiKey(ctx context.Context, apiKey string) (dbqueries.GetApiKeyRow, error)
}

// RequestAuthenticator authenticates a request with a key, returning an
// auth context that represents the authenticated entity.
type RequestAuthenticator interface {
	Authenticate(ctx context.Context, key string) (AuthContext, error)
}

// AuthRequire validates that incoming requests provide either a valid Api-Key or a valid Bearer
// JWT.
func AuthRequire(
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		log.Printf("auth require")
		apiKey := strings.TrimSpace(ginContext.GetHeader("Api-Key"))
		authorizationValue := strings.TrimSpace(ginContext.GetHeader("Authorization"))

		authContext, err := authenticateRequest(
			ginContext.Request.Context(),
			jwtAuthenticator,
			apiKeyAuthenticator,
			apiKey,
			authorizationValue,
		)
		if err == nil {
			ginContext.Set(authContextContextKey, authContext)
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

// GetAuthContext retrieves the authenticated request context from Gin context.
func GetAuthContext(ginContext *gin.Context) (AuthContext, bool) {
	rawPrincipal, exists := ginContext.Get(authContextContextKey)
	if !exists {
		return AuthContext{Scheme: "", UserID: "", APIKey: nil}, false
	}

	principal, ok := rawPrincipal.(AuthContext)
	if !ok {
		return AuthContext{Scheme: "", UserID: "", APIKey: nil}, false
	}

	return principal, true
}

// EffectiveUserID returns the acting user ID for requests authenticated via bearer auth
// or via a user-owned API key. It returns false for service-owned API keys and requests
// without an authenticated user identity.
func EffectiveUserID(authContext AuthContext) (string, bool) {
	if authContext.Scheme == AuthSchemeBearer && authContext.UserID != "" {
		return authContext.UserID, true
	}

	if authContext.Scheme != AuthSchemeAPIKey || authContext.APIKey == nil {
		return "", false
	}

	if authContext.APIKey.PrincipalType != dbqueries.PrincipalTypeUser ||
		authContext.APIKey.UserID == nil {
		return "", false
	}

	return *authContext.APIKey.UserID, true
}

// GetEffectiveUserID returns the acting user ID for requests authenticated via bearer auth
// or via a user-owned API key. It returns false for service-owned API keys and requests
// without an authenticated user identity.
func GetEffectiveUserID(ginContext *gin.Context) (string, bool) {
	authContext, ok := GetAuthContext(ginContext)
	if !ok {
		return "", false
	}

	return EffectiveUserID(authContext)
}

// authenticateRequest authenticates an incoming request. Supports both API key and
// bearer JWT authentication.
//
// The returned AuthContext provides context on what scheme was used to authenticate the
// request (and additional context, such as the user id in the case of bearer JWT auth).
//
//nolint:cyclop // This function intentionally implements fallback auth scheme evaluation.
func authenticateRequest(
	ctx context.Context,
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
	apiKey string,
	authorizationValue string,
) (AuthContext, error) {
	var internalErrors []error

	var invalidCredentialErrors []error

	if authorizationValue != "" {
		rawToken, err := bearerTokenFromHeader(authorizationValue)
		switch {
		case err == nil:
			authContext, verifyErr := jwtAuthenticator.Authenticate(ctx, rawToken)
			switch {
			case verifyErr == nil:
				return authContext, nil
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
		authContext, err := apiKeyAuthenticator.Authenticate(ctx, apiKey)
		switch {
		case err == nil:
			return authContext, nil
		case errors.Is(err, errInvalidAPIKey):
			invalidCredentialErrors = append(invalidCredentialErrors, err)
		default:
			internalErrors = append(internalErrors, err)
		}
	}

	if len(internalErrors) > 0 {
		return AuthContext{}, errors.Join(internalErrors...)
	}

	if len(invalidCredentialErrors) > 0 {
		return AuthContext{}, errors.Join(invalidCredentialErrors...)
	}

	return AuthContext{Scheme: "", UserID: "", APIKey: nil}, errMissingCredentials
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
		Message: &message,
	})
}
