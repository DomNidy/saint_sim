// Package main runs the Gin API server.
package main

import (
	"context"
	"log"
	"net/http"

	gin "github.com/gin-gonic/gin"

	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	middleware "github.com/DomNidy/saint_sim/apps/api/middleware"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // include filename:linenumber in log output

	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	simulationQueue, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)

		return
	}

	pool := utils.InitPostgresConnectionPool(context.Background())

	dbClient := dbqueries.New(pool)
	if dbClient == nil {
		log.Panicf("API failed to acquire a dbClient, cannot run.")
	}

	defer pool.Close()
	defer simulationQueue.Close()

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
	authorized := router.Group("/", middleware.AuthRequire(*dbClient))

	authorized.POST("/simulation", func(ginContext *gin.Context) {
		handlers.Simulate(ginContext, dbClient, simulationQueue)
	})

	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Printf("ERROR: Failed to start API server: %v", err)
	}
}
