package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/apps/api/repositories"
	"github.com/DomNidy/saint_sim/pkg/interfaces"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
	gin "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	// todo: implement auth
	r.POST("/simulate", func(c *gin.Context) {
		var simOptions interfaces.SimulationOptions
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
		simulationMessageBody, err := json.Marshal(interfaces.SimulationMessageBody{
			SimulationId: &simulationRequestId,
		})
		if err != nil {
			log.Printf("Error converting to json: %v", receivedJson)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		log.Printf("Got simulation request options: %s", string(receivedJson))
		log.Printf("Marshalled SimulationMessageBody into JSON object: %s", string(simulationMessageBody))

		// Write operation request to db
		_, err = db.Exec(context.Background(), "INSERT INTO simulation_request (id, options) VALUES ($1, $2)", simulationRequestId, simOptions)

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
			SimulationRequestId: utils.StrPtr(string(simulationRequestId[:])),
		})
	})
	r.GET("/simulate/:id", func(c *gin.Context) {
		// get sim id from params & convert to int
		simulationIdStr, _ := c.Params.Get("id")
		simulationId, err := strconv.Atoi(simulationIdStr)
		if err != nil {
			log.Printf("%v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid simulation id"})
			return
		}

		repo := repositories.NewSimDataRepository(db)
		data, err := repo.GetSimData(simulationId)
		if err != nil {
			log.Printf("error getting sim data:", err)
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Could not find simulation data with this id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, data)

	})
	r.Run("0.0.0.0:8080")
}
