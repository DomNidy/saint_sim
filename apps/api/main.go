// Package main runs the Gin API server.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	gin "github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4/jwt"

	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	middleware "github.com/DomNidy/saint_sim/apps/api/middleware"
	dbqueries "github.com/DomNidy/saint_sim/pkg/db"
	"github.com/DomNidy/saint_sim/pkg/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
)

func newJWTAuthenticator(
	dbClient *dbqueries.Queries,
	betterAuthURL string,
) *middleware.JWTAuthenticator {
	return middleware.NewJWTAuthenticator(dbClient, &jwt.Expected{
		// only accept jwt issued by our web app
		Issuer: betterAuthURL,
		// dont accept allow tokens for other audiences
		AnyAudience: []string{"saint-api"},
		Subject:     "",
		ID:          "",
		Time:        time.Time{},
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // include filename:linenumber in log output

	betterAuthURL := secrets.LoadSecret("BETTER_AUTH_URL").Value()
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	simulationQueue, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)

		return
	}
	defer simulationQueue.Close()

	pool, err := utils.InitPostgresConnectionPool(context.Background())
	if err != nil {
		log.Panicf("%s: could not make postgres pool", err)
	}
	defer pool.Close()

	dbClient := dbqueries.New(pool)
	if dbClient == nil {
		log.Panicf("API failed to acquire a dbClient, cannot run.")
	}
	// Setup api server
	router := gin.Default()
	router.GET("/health", func(ginContext *gin.Context) {
		ginContext.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	router.GET("/simulation/:id", func(ginContext *gin.Context) {
		handlers.GetSimulation(ginContext, dbClient)
	})

	// Authorization group: https://gin-gonic.com/zh-tw/docs/examples/using-middleware/
	apiKeyAuthenticator := middleware.NewAPIKeyAuthenticator(dbClient)
	jwtAuthenticator := newJWTAuthenticator(dbClient, betterAuthURL)
	authorized := router.Group("/", middleware.AuthRequire(jwtAuthenticator, apiKeyAuthenticator))

	authorized.POST("/simulation", func(ginContext *gin.Context) {
		handlers.Simulate(ginContext, dbClient, simulationQueue)
	})

	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Printf("ERROR: Failed to start API server: %v", err)
	}
}
