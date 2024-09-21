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
		// todo: this should return an error if the request json body does not match the SimulationOptions type definition
		if err := c.ShouldBindJSON(&simOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create SimulationMessageBody
		simulationId := api_utils.GenerateUUID()
		simulationMessageBody, err := json.Marshal(interfaces.SimulationMessageBody{
			SimulationId: &simulationId,
		})
		utils.FailOnError(err, "Failed to marshal SimulationMessageBody into JSON")
		log.Printf("Marshalled SimulationMessageBody into JSON object: %s", string(simulationMessageBody))

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
