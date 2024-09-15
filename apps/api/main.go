package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	secrets "github.com/DomNidy/saint_sim/pkg/secrets"
	gin "github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

// TODO: /health endpoint openapi definition
// TODO: /simulate endpoint impl

func main() {
	RABBITMQ_USER := secrets.LoadSecret("RABBITMQ_USER")
	RABBITMQ_PASS := secrets.LoadSecret("RABBITMQ_PASS")
	RABBITMQ_PORT := secrets.LoadSecret("RABBITMQ_PORT")
	RABBITMQ_HOST := secrets.LoadSecret("RABBITMQ_HOST")
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s:%s", RABBITMQ_USER.Value(), RABBITMQ_PASS.Value(), RABBITMQ_HOST.Value(), RABBITMQ_PORT.Value())
	fmt.Printf("%s\n", connectionURI)
	// Connect to rabbitmq
	conn, err := amqp.Dial(connectionURI)

	failOnError(err, "Failed to connect to rabbitmq")
	defer conn.Close()

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

		if err := c.ShouldBindJSON(&simOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("Got sim options:")
		fmt.Printf("char name: %s\n", *simOptions.WowCharacter.CharacterName)
		fmt.Printf("realm name: %s\n", *simOptions.WowCharacter.Realm)
		fmt.Printf("regions name: %s\n", *simOptions.WowCharacter.Region)

		// todo: post to rabbitmq
	})
	r.Run("0.0.0.0:8080")
}
