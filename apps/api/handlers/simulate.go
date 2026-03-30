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
)

type wowCharacterValidationFailure struct {
	statusCode int
	response   gin.H
}

func Simulate(c *gin.Context, dbClient *db.Queries, simQueue *utils.SimulationQueueClient) {
	var simOptions api_types.SimulationOptions
	if err := c.ShouldBindJSON(&simOptions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid simulation options"})
		return
	}

	// Validate sim options to prevent rce
	if !utils.IsValidSimOptions(&simOptions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	if validationFailure, err := validateWowCharacter(simOptions.WowCharacter); err != nil {
		log.Printf("Error checking for character existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	} else if validationFailure != nil {
		c.JSON(validationFailure.statusCode, validationFailure.response)
		return
	}

	simOptionsJSON, err := json.Marshal(simOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create the simulation entry in the DB
	simEntry, err := dbClient.CreateSimulation(context.Background(), simOptionsJSON)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	idValue, err := simEntry.ID.Value()
	if err != nil {
		log.Printf("Error converting simulation id to string: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	simulationID, ok := idValue.(string)
	if !ok {
		log.Printf("Unexpected simulation id type: %T", idValue)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	simulationMessageBody := utils.SimulationMessage{
		SimulationID: simulationID,
	}

	err = simQueue.Publish(simulationMessageBody)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)
		c.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("An internal server error occurred. Please try again later."),
		})
		return
	}

	log.Printf(" [x] Sent %v\n", simulationMessageBody)
	c.JSON(200, api_types.Simulation{
		Id:     &simulationMessageBody.SimulationID,
		Status: (*api_types.SimulationStatus)(utils.StrPtr("in_queue")),
	})

}

func validateWowCharacter(character api_types.WowCharacter) (*wowCharacterValidationFailure, error) {
	if !utils.IsValidWowRealm(string(character.Realm)) {
		return &wowCharacterValidationFailure{
			statusCode: http.StatusBadRequest,
			response:   gin.H{"error": "Invalid wow realm"},
		}, nil
	}

	if !utils.IsValidWowRegion(string(character.Region)) {
		return &wowCharacterValidationFailure{
			statusCode: http.StatusBadRequest,
			response:   gin.H{"error": "Invalid wow region"},
		}, nil
	}

	exists, err := api_utils.CheckWowCharacterExists(&character)
	if err != nil {
		return nil, err
	}
	if !exists {
		log.Printf("WoW character does not exist")
		return &wowCharacterValidationFailure{
			statusCode: http.StatusNotFound,
			response:   gin.H{"message": "WoW character does not exist"},
		}, nil
	}

	return nil, nil
}
