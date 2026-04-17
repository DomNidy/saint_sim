package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

var (
	errMissingRequestValidationInput = errors.New("missing request validation input")
	errUnsupportedSecurityScheme     = errors.New("unsupported security scheme")
)

type APIAuthenticator struct {
}

// AuthenticateOpenAPIRequest authenticates a request for a single OpenAPI
// security scheme. The resolved AuthContext will be set on the request context.
//
// The server's oapi middleware will call this for each security requirement on a
// route, until one passes (doesn't return an error). The `input` here encodes the
// security requirement that we are validating against.
//
// Example: openapi definition for GET /pets route defines `BearerAuth` and `ApiKeyAuth`,
// For a single request:
// AuthenticateOpenAPIRequest call 1: try validate BearerAuth scheme.
// if failed, continue:
// AuthenticateOpenAPIRequest call 2: try validate ApiKeyAuth scheme.
// etc.
func AuthenticateOpenAPIRequest(
	ctx context.Context,
	input *openapi3filter.AuthenticationInput,
	jwtAuthenticator RequestAuthenticator,
	apiKeyAuthenticator RequestAuthenticator,
) error {
	if input == nil ||
		input.RequestValidationInput == nil ||
		input.RequestValidationInput.Request == nil {
		return errMissingRequestValidationInput
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
		return fmt.Errorf("openapi authentication failed: %w", input.NewError(err))
	}

	ginContext := ginmiddleware.GetGinContext(ctx)
	if ginContext != nil {
		SetAuthContext(ginContext, authContext)
	}

	return nil
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
