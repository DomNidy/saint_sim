// Package middleware wires shared HTTP middleware for the API server.
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	"github.com/DomNidy/saint_sim/internal/api"
)

const authorizationHeaderSplitParts = 2

// OpenAPIValidation returns a handler that validates requests against the
// embedded swagger spec and performs per-scheme authentication for secured
// operations.
func OpenAPIValidation(
	swagger *openapi3.T,
	authenticateRequest auth.OpenAPIRequestAuthenticator,
) gin.HandlerFunc {
	return ginmiddleware.OapiRequestValidatorWithOptions(swagger, &ginmiddleware.Options{
		ErrorHandler: openAPIErrorHandler,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(
				ctx context.Context,
				input *openapi3filter.AuthenticationInput,
			) error {
				if authenticateRequest != nil {
					authContext, err := authenticateRequest(ctx, input)
					if err != nil {
						return err
					}

					ginContext := ginmiddleware.GetGinContext(ctx)
					if ginContext != nil {
						auth.SetAuthContext(ginContext, authContext)
					}

					return nil
				}

				log.Printf(
					"WARNING: EVERYTHING IS PUBLIC !!! - no AuthenticationFunc was provided to openapi" +
						"validation middleware - EVERYTHING IS PUBLIC!!!",
				)

				// return nil == request authenticated.
				return nil
			},
			ExcludeRequestBody:          false,
			ExcludeRequestQueryParams:   false,
			ExcludeResponseBody:         false,
			ExcludeReadOnlyValidations:  false,
			ExcludeWriteOnlyValidations: false,
			IncludeResponseStatus:       false,
			MultiError:                  false,
			SkipSettingDefaults:         false,
		},
		MultiErrorHandler:     nil,
		ParamDecoder:          nil,
		SilenceServersWarning: false,
		UserData:              nil,
	})
}

func openAPIErrorHandler(ginContext *gin.Context, message string, statusCode int) {
	if strings.Contains(message, "openapi3filter.SecurityRequirementsError") {
		logUnauthorizedRequest(ginContext, message)
		abortWithError(ginContext, http.StatusUnauthorized, "Unauthorized")

		return
	}

	if statusCode == http.StatusNotFound {
		abortWithError(ginContext, statusCode, "Not found")

		return
	}

	abortWithError(ginContext, statusCode, message)
}

func logUnauthorizedRequest(ginContext *gin.Context, message string) {
	if ginContext == nil || ginContext.Request == nil {
		log.Printf("unauthorized request: message=%q", message)

		return
	}

	request := ginContext.Request
	authorizationValue := strings.TrimSpace(request.Header.Get("Authorization"))
	apiKey := strings.TrimSpace(request.Header.Get("Api-Key"))
	authorizationScheme := ""
	if authorizationValue != "" {
		authorizationScheme = strings.TrimSpace(
			strings.SplitN(authorizationValue, " ", authorizationHeaderSplitParts)[0],
		)
	}

	log.Printf(
		"unauthorized request:"+
			" method=%s path=%s client_ip=%s"+
			" authorization_present=%t authorization_scheme=%q"+
			" api_key_present=%t api_key_length=%d"+
			" user_agent=%q auth_error=%q",
		request.Method,
		request.URL.Path,
		ginContext.ClientIP(),
		authorizationValue != "",
		authorizationScheme,
		apiKey != "",
		len(apiKey),
		request.UserAgent(),
		message,
	)
}

func abortWithError(ginContext *gin.Context, statusCode int, message string) {
	ginContext.AbortWithStatusJSON(statusCode, api.ErrorResponse{
		Message: message,
		Code:    "",
	})
}
