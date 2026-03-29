package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	"github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func Simulate(c *gin.Context, dbClient *db.Queries, simQueue *utils.SimulationQueueClient) {
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
	simulationMessageBody := api_types.SimulationMessageBody{
		SimulationId: &simulationRequestId,
	}
	simulationMessageBodyJson, err := json.Marshal(simulationMessageBody)
	if err != nil {
		log.Printf("Error converting to json: %v, %v", receivedJson, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("Got simulation request options: %s", string(receivedJson))
	log.Printf("Marshalled SimulationMessageBody into JSON object: %s", string(simulationMessageBodyJson))

	var simulationRequestUUID pgtype.UUID
	if err := simulationRequestUUID.Scan(simulationRequestId); err != nil {
		log.Printf("Error converting simulation request id to uuid: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	simRequestParams := db.CreateSimulationRequestParams{
		ID:      simulationRequestUUID,
		Options: receivedJson,
	}
	// Create the simulation_request entry in the DB
	// TODO: We can add request origin information here so we can later determine if we should trigger the discord notification postgres channel
	// TODO: We will need to update the 'notify_simulation_data' trigger in the db to check for this info.
	err = dbClient.CreateSimulationRequest(context.Background(), simRequestParams)
	if err != nil {
		log.Printf("%v", err)
	}

	err = simQueue.Publish(simulationMessageBody)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)
		c.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("An internal server error occurred. Please try again later."),
		})
		return
	}

	log.Printf(" [x] Sent %s\n", simulationMessageBody)
	c.JSON(200, api_types.SimulationResponse{
		SimulationRequestId: utils.StrPtr(string(simulationRequestId[:])),
	})
}
