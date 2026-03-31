package main

import (
	"context"
	"log"

	gin "github.com/gin-gonic/gin"

	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	middleware "github.com/DomNidy/saint_sim/apps/api/middleware"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

var queue *utils.SimulationQueueClient

func init() {
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	q, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)
		return
	}

	queue = q
}

func main() {
	pool := utils.InitPostgresConnectionPool(context.Background())
	db := dbqueries.New(pool)

	defer pool.Close()
	defer queue.Close()

	// Setup api server
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Authorization group: https://gin-gonic.com/zh-tw/docs/examples/using-middleware/
	authorized := r.Group("/", func(ctx *gin.Context) { middleware.AuthRequire(db) })

	authorized.POST("/simulate", func(ctx *gin.Context) { handlers.Simulate(ctx, db, queue) })

	r.Run("0.0.0.0:8080")
}
