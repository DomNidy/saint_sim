package main

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	"github.com/DomNidy/saint_sim/apps/api/middleware"
	api "github.com/DomNidy/saint_sim/internal/api"
)

func newRouter(
	server api.StrictServerInterface,
	swagger *openapi3.T,
	jwtAuthenticator auth.RequestAuthenticator,
	apiKeyAuthenticator auth.RequestAuthenticator,
) *gin.Engine {
	router := gin.Default()

	authFunc := openapi3filter.AuthenticationFunc(
		func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
			return auth.AuthenticateOpenAPIRequest(
				ctx,
				input,
				jwtAuthenticator,
				apiKeyAuthenticator,
			)
		},
	)

	apiRouter := router.Group("/")
	apiRouter.Use(middleware.OpenAPIValidation(swagger, &authFunc))

	api.RegisterHandlersWithOptions(
		apiRouter,
		api.NewStrictHandler(server, nil),
		api.GinServerOptions{
			ErrorHandler: nil,
			BaseURL:      "",
			Middlewares:  nil,
		},
	)

	return router
}
