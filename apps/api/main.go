package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/apps/api/handlers"
	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	gin "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	db := utils.InitPostgresConnectionPool(context.Background())
	defer db.Close()
	queries := dbqueries.New(db)

	conn, ch := utils.InitRabbitMQConnection()
	defer conn.Close()
	defer ch.Close()

	// Declare queue to publish msgs to
	q := utils.DeclareSimulationQueue(ch)

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

		// Validate sim options to prevent rce
		if !utils.IsValidSimOptions(&simOptions) {
			log.Printf("Sim options were invalid!")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		// Marshal back
		receivedJson, err := json.Marshal(simOptions)
		if err != nil {
			log.Printf("Error converting to json: %v, %v", receivedJson, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Make sure the provided realm & regions exist
		if !utils.IsValidWowRealm(string(simOptions.WowCharacter.Realm)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wow realm"})
			return
		}

		if !utils.IsValidWowRegion(string(simOptions.WowCharacter.Region)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wow region"})
			return
		}

		// Make sure the wow character actually exists before sending sim msg
		exists, err := api_utils.CheckWowCharacterExists(&simOptions.WowCharacter)
		if err != nil {
			log.Printf("Error checking for character existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		} else if !exists {
			log.Printf("WoW character does not exist")
			c.JSON(http.StatusNotFound, gin.H{"message": "WoW character does not exist"})
			return
		}

		// Create SimulationMessageBody
		simulationRequestId := api_utils.GenerateUUID()
		simulationMessageBody, err := json.Marshal(api_types.SimulationMessageBody{
			SimulationId: &simulationRequestId,
		})
		if err != nil {
			log.Printf("Error converting to json: %v, %v", receivedJson, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		log.Printf("Got simulation request options: %s", string(receivedJson))
		log.Printf("Marshalled SimulationMessageBody into JSON object: %s", string(simulationMessageBody))

		var simulationRequestUUID pgtype.UUID
		if err := simulationRequestUUID.Scan(simulationRequestId); err != nil {
			log.Printf("Error converting simulation request id to uuid: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Create the simulation_request entry in the DB
		// TODO: We can add request origin information here so we can later determine if we should trigger the discord notification postgres channel
		// TODO: We will need to update the 'notify_simulation_data' trigger in the db to check for this info.
		err = queries.CreateSimulationRequest(context.Background(), dbqueries.CreateSimulationRequestParams{
			ID:      simulationRequestUUID,
			Options: receivedJson,
		})

		if err != nil {
			log.Printf("%v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        simulationMessageBody,
			},
		)
		utils.FailOnError(err, "Failed to publish msg to queue")
		log.Printf(" [x] Sent %s\n", simulationMessageBody)
		c.JSON(200, api_types.SimulationResponse{
			SimulationRequestId: utils.StrPtr(string(simulationRequestId[:])),
		})
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
	r.Run("0.0.0.0:8080")
}
