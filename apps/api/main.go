package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	middleware "github.com/DomNidy/saint_sim/apps/api/middleware"
	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	gin "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var queue *utils.SimulationQueueClient = nil

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

	r.GET("/report/:id", func(c *gin.Context) {
		// get sim id from params & convert to int
		simulationIdStr, _ := c.Params.Get("id")
		simulationId, err := strconv.Atoi(simulationIdStr)
		if err != nil {
			log.Printf("%v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid simulation id"})
			return
		}

		simData, err := db.GetSimulationData(c, int32(simulationId))
		if err != nil {
			log.Printf("error getting sim data: %v", err.Error())
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Could not find simulation data with this id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, api_types.SimulationData{
			Id:        &simulationId,
			RequestId: utils.StrPtr(string(simData.RequestID.Bytes[:])),
			SimResult: utils.StrPtr(simData.SimResult),
		})

	})
	r.Run("0.0.0.0:8080")
}
