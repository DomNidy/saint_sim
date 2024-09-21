package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	q, err := ch.QueueDeclare(
		"simulation_queue", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                //arguments
	)
	utils.FailOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup api server
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})
	// todo: continue implementation
	r.POST("/simulate", func(c *gin.Context) {
		var simOptions interfaces.SimulateJSONRequestBody

		if err := c.ShouldBindJSON(&simOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("Got sim options:")
		fmt.Printf("char name: %s\n", *simOptions.WowCharacter.CharacterName)
		fmt.Printf("realm name: %s\n", *simOptions.WowCharacter.Realm)
		fmt.Printf("regions name: %s\n", *simOptions.WowCharacter.Region)

		msgBodyJson, err := json.Marshal(simOptions)
		utils.FailOnError(err, "Failed to marshal json")

		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        msgBodyJson,
			},
		)
		utils.FailOnError(err, "Failed to publish msg to queue")
		fmt.Printf(" [x] Sent %s\n", msgBodyJson)

		c.JSON(200, interfaces.SimulationResponse{
			SimulationId: utils.StrPtr("some_id_here"),
		})
	})
	r.Run("0.0.0.0:8080")
}
