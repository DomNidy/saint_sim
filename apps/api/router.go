package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/apps/api/middleware"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/utils"
)

func newRouter(
	server api.StrictServerInterface,
	swagger *openapi3.T,
	jwtAuthenticator middleware.RequestAuthenticator,
	apiKeyAuthenticator middleware.RequestAuthenticator,
) *gin.Engine {
	router := gin.Default()

	apiRouter := router.Group("/")
	apiRouter.Use(middleware.OpenAPIValidation(swagger, jwtAuthenticator, apiKeyAuthenticator))

	api.RegisterHandlersWithOptions(
		apiRouter,
		api.NewStrictHandler(server, nil),
		api.GinServerOptions{
			BaseURL:      "",
			ErrorHandler: generatedRouteErrorHandler,
			Middlewares:  nil,
		},
	)

	return router
}

func generatedRouteErrorHandler(ginContext *gin.Context, err error, statusCode int) {
	ginContext.AbortWithStatusJSON(statusCode, api.ErrorResponse{
		Message: utils.StrPtr(err.Error()),
	})
}
