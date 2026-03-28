package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/apps/api/handlers"
	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	gin "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func Run() error {
	db := utils.InitPostgresConnectionPool(context.Background())
	defer db.Close()
	queries := dbqueries.New(db)

	conn, ch := utils.InitRabbitMQConnection()
	defer conn.Close()
	defer ch.Close()

	// Declare queue to publish msgs to
	q := utils.DeclareSimulationQueue(ch)

	simulationService := NewSimulationService(
		queries,
		rabbitMQDispatcher{
			channel:   ch,
			queueName: q.Name,
			timeout:   5 * time.Second,
		},
		liveCharacterLookup{},
		api_utils.GenerateUUID,
	)

	// Setup api server
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Authorization group: https://gin-gonic.com/zh-tw/docs/examples/using-middleware/
	authorized := r.Group("/", handlers.AuthRequire(db))

	authorized.POST("/simulate", func(c *gin.Context) {
		var simOptions api_types.SimulationOptions
		if err := c.ShouldBindJSON(&simOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid simulation options"})
			return
		}

		response, err := simulationService.Submit(c, simOptions)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidSimOptions):
				log.Printf("Sim options were invalid")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			default:
				log.Printf("simulate request failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			return
		}

		c.JSON(http.StatusOK, response)
	})
	r.GET("/report/:id", func(c *gin.Context) {
		// get sim id from params & convert to int
		simulationIdStr, _ := c.Params.Get("id")
		simulationId, err := strconv.Atoi(simulationIdStr)
		if err != nil {
			log.Printf("%v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid simulation id"})
			return
		}

		simData, err := queries.GetSimulationData(c, int32(simulationId))
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
	return r.Run("0.0.0.0:8080")
}
