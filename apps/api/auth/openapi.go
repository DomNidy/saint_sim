package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
)

var (
	errMissingRequestValidationInput = errors.New("missing request validation input")
	errUnsupportedSecurityScheme     = errors.New("unsupported security scheme")
)

// OpenAPIRequestAuthenticator authenticates a single OpenAPI security
// requirement and returns the resolved request auth context.
type OpenAPIRequestAuthenticator func(
	ctx context.Context,
	input *openapi3filter.AuthenticationInput,
) (AuthContext, error)

// NewOpenAPIRequestAuthenticator wires the core request authenticators into an
// OpenAPI security requirement authenticator that can be consumed by middleware.
func NewOpenAPIRequestAuthenticator(
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
) OpenAPIRequestAuthenticator {
	return func(
		ctx context.Context,
		input *openapi3filter.AuthenticationInput,
	) (AuthContext, error) {
		return authenticateOpenAPIRequest(ctx, input, jwtAuthenticator, apiKeyAuthenticator)
	}
}

// authenticateOpenAPIRequest authenticates a request for a single OpenAPI
// security scheme and returns the resolved auth context.
//
// The server's oapi middleware will call this for each security requirement on a
// route, until one passes (doesn't return an error). The `input` here encodes the
// security requirement that we are validating against.
//
// Example: openapi definition for GET /pets route defines `BearerAuth` and `ApiKeyAuth`,
// For a single request:
// authenticateOpenAPIRequest call 1: try validate BearerAuth scheme.
// if failed, continue:
// authenticateOpenAPIRequest call 2: try validate ApiKeyAuth scheme.
// etc.
func authenticateOpenAPIRequest(
	ctx context.Context,
	input *openapi3filter.AuthenticationInput,
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
) (AuthContext, error) {
	if input == nil ||
		input.RequestValidationInput == nil ||
		input.RequestValidationInput.Request == nil {
		return AuthContext{}, errMissingRequestValidationInput
	}

	request := input.RequestValidationInput.Request
	authContext, err := authenticateOpenAPISecurityRequirement(
		ctx,
		input.SecuritySchemeName,
		request.Header.Get("Authorization"),
		request.Header.Get("Api-Key"),
		jwtAuthenticator,
		apiKeyAuthenticator,
	)
	if err != nil {
		return AuthContext{}, fmt.Errorf("openapi authentication failed: %w", input.NewError(err))
	}

	return authContext, nil
}

func authenticateOpenAPISecurityRequirement(
	ctx context.Context,
	securitySchemeName string,
	authorizationValue string,
	apiKey string,
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
) (AuthContext, error) {
	switch securitySchemeName {
	case "BearerAuth":
		rawAuthorization := strings.TrimSpace(authorizationValue)
		if rawAuthorization == "" {
			return AuthContext{}, errMissingCredentials
		}

		token, err := bearerTokenFromHeader(rawAuthorization)
		if err != nil {
			return AuthContext{}, err
		}

		authContext, authErr := jwtAuthenticator.Authenticate(ctx, token)
		if authErr != nil {
			return AuthContext{}, fmt.Errorf("authenticate bearer token: %w", authErr)
		}

		return authContext, nil
	case "ApiKeyAuth":
		rawAPIKey := strings.TrimSpace(apiKey)
		if rawAPIKey == "" {
			return AuthContext{}, errMissingCredentials
		}

		authContext, authErr := apiKeyAuthenticator.Authenticate(ctx, rawAPIKey)
		if authErr != nil {
			return AuthContext{}, fmt.Errorf("authenticate api key: %w", authErr)
		}

		return authContext, nil
	default:
		return AuthContext{}, errUnsupportedSecurityScheme
	}
}
