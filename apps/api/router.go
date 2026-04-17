package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	"github.com/DomNidy/saint_sim/apps/api/middleware"
	api "github.com/DomNidy/saint_sim/internal/api"
)

func newRouter(
	server api.StrictServerInterface,
	swagger *openapi3.T,
	openAPIAuthenticator auth.OpenAPIRequestAuthenticator,
) *gin.Engine {
	router := gin.Default()

	apiRouter := router.Group("/")
	apiRouter.Use(middleware.OpenAPIValidation(swagger, openAPIAuthenticator))

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
