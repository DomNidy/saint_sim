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
		abortWithError(ginContext, http.StatusUnauthorized, "Unauthorized")

		return
	}

	if statusCode == http.StatusNotFound {
		abortWithError(ginContext, statusCode, "Not found")

		return
	}

	abortWithError(ginContext, statusCode, message)
}

func abortWithError(ginContext *gin.Context, statusCode int, message string) {
	ginContext.AbortWithStatusJSON(statusCode, api.ErrorResponse{
		Message: &message,
	})
}
