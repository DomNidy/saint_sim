package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/pkg/interfaces"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
	gin "github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO: /health endpoint openapi definition
// TODO: /simulate endpoint impl

func main() {
	db := utils.InitPostgresConnectionPool(context.Background())
	defer db.Close()

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
	// todo: continue implementation
	r.POST("/simulate", func(c *gin.Context) {
		var simOptions interfaces.SimulationOptions

		// bind and validate json from request
		// todo: this should return an error if the request json body does not match the SimulationOptions type definition
		if err := c.ShouldBindJSON(&simOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate sim options to prevent rce
		if !utils.IsValidSimOptions(&simOptions) {
			log.Printf("Sim options were invalid!")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		receivedJson, err := json.Marshal(simOptions)
		if err != nil {
			log.Printf("Error converting to json: %v", receivedJson)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Make sure the wow character actually exists before sending sim msg
		exists, err := api_utils.CheckWowCharacterExists(simOptions.WowCharacter)
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
		simulationId := api_utils.GenerateUUID()
		simulationMessageBody, err := json.Marshal(interfaces.SimulationMessageBody{
			SimulationId: &simulationId,
		})
		if err != nil {
			log.Printf("Error converting to json: %v", receivedJson)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		log.Printf("Got simulation request options: %s", string(receivedJson))
		log.Printf("Marshalled SimulationMessageBody into JSON object: %s", string(simulationMessageBody))

		// Write operation request to db
		_, err = db.Exec(context.Background(), "INSERT INTO simulation_request (id, options) VALUES ($1, $2)", simulationId, simOptions)

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
		c.JSON(200, interfaces.SimulationResponse{
			SimulationId: utils.StrPtr(string(simulationId[:])),
		})
	})
	r.Run("0.0.0.0:8080")
}
