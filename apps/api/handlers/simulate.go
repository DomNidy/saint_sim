// Package handlers contains the Gin handlers for the API server.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	"github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

type wowCharacterValidationError struct {
	statusCode int
	response   api_types.ErrorResponse
}

func (failure *wowCharacterValidationError) Error() string {
	if failure == nil || failure.response.Message == nil {
		return "simulation validation failed"
	}

	return *failure.response.Message
}

// Simulate validates a simulation request, persists it, and enqueues it for processing.
func Simulate(
	ginContext *gin.Context,
	dbClient *db.Queries,
	simQueue *utils.SimulationQueueClient,
) {
	simOptions, err := decodeAndValidateSimulationRequest(ginContext)
	if err != nil {
		var validationFailure *wowCharacterValidationError
		if errors.As(err, &validationFailure) {
			ginContext.JSON(validationFailure.statusCode, validationFailure.response)

			return
		}

		log.Printf("Error checking for character existence: %v", err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("Internal server error"),
		})

		return
	}

	simulationID, err := createSimulationRequest(ginContext, dbClient, simOptions)
	if err != nil {
		log.Printf("Error creating simulation request: %v", err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("Internal server error"),
		})

		return
	}

	simulationMessageBody := utils.SimulationMessage{
		SimulationID: simulationID,
	}

	err = simQueue.Publish(simulationMessageBody)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("An internal server error occurred. Please try again later."),
		})

		return
	}

	log.Printf(" [x] Sent %v\n", simulationMessageBody)
	ginContext.JSON(http.StatusAccepted, gin.H{
		"simulation_id": simulationMessageBody.SimulationID,
	})
}

func decodeAndValidateSimulationRequest(c *gin.Context) (api_types.SimulationOptions, error) {
	var simOptions api_types.SimulationOptions

	err := c.ShouldBindJSON(&simOptions)
	if err != nil {
		return api_types.SimulationOptions{}, &wowCharacterValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Invalid simulation options"),
			},
		}
	}

	if !utils.IsValidSimOptions(&simOptions) {
		return api_types.SimulationOptions{}, &wowCharacterValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Bad request"),
			},
		}
	}

	err = validateWowCharacter(simOptions.WowCharacter)
	if err != nil {
		return api_types.SimulationOptions{}, err
	}

	return simOptions, nil
}

func createSimulationRequest(
	ginContext *gin.Context,
	dbClient *db.Queries,
	simOptions api_types.SimulationOptions,
) (string, error) {
	simOptionsJSON, err := json.Marshal(simOptions)
	if err != nil {
		return "", fmt.Errorf("marshal simulation options: %w", err)
	}

	simEntry, err := dbClient.CreateSimulation(ginContext.Request.Context(), simOptionsJSON)
	if err != nil {
		return "", fmt.Errorf("create simulation row: %w", err)
	}

	simulationID, err := utils.UUIDString(simEntry.ID)
	if err != nil {
		return "", fmt.Errorf("convert simulation id to string: %w", err)
	}

	return simulationID, nil
}

// GetSimulation returns the current state or completed result for a simulation.
func GetSimulation(ginContext *gin.Context, dbClient *db.Queries) {
	var simulationID pgtype.UUID

	err := simulationID.Scan(ginContext.Param("id"))
	if err != nil {
		ginContext.JSON(http.StatusNotFound, api_types.ErrorResponse{
			Message: utils.StrPtr("Simulation not found"),
		})

		return
	}

	simulation, err := dbClient.GetSimulation(ginContext.Request.Context(), simulationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ginContext.JSON(http.StatusNotFound, api_types.ErrorResponse{
				Message: utils.StrPtr("Simulation not found"),
			})

			return
		}

		log.Printf("Error loading simulation %s: %v", ginContext.Param("id"), err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("Internal server error"),
		})

		return
	}

	response, err := simulationResponseFromRecord(simulation)
	if err != nil {
		log.Printf("Error serializing simulation %s: %v", ginContext.Param("id"), err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("Internal server error"),
		})

		return
	}

	ginContext.JSON(http.StatusOK, response)
}

func simulationResponseFromRecord(simulation db.Simulation) (api_types.Simulation, error) {
	simulationID, err := utils.UUIDString(simulation.ID)
	if err != nil {
		return api_types.Simulation{}, fmt.Errorf("convert simulation id to string: %w", err)
	}

	var response api_types.Simulation

	status := simulationStatusFromRecord(simulation)

	response.Id = &simulationID
	response.SimulationStatus = &status

	if simulation.SimResult.Valid {
		response.SimResult = &simulation.SimResult.String
	}

	if simulation.ErrorText.Valid {
		response.ErrorText = &simulation.ErrorText.String
	}

	return response, nil
}

func simulationStatusFromRecord(simulation db.Simulation) api_types.SimulationStatus {
	if simulation.ErrorText.Valid {
		return api_types.Error
	}

	if simulation.CompletedAt.Valid {
		return api_types.Complete
	}

	if simulation.StartedAt.Valid {
		return api_types.InProgress
	}

	return api_types.InQueue
}

func validateWowCharacter(character api_types.WowCharacter) error {
	if !utils.IsValidWowRealm(string(character.Realm)) {
		return &wowCharacterValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Invalid wow realm"),
			},
		}
	}

	if !utils.IsValidWowRegion(string(character.Region)) {
		return &wowCharacterValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Invalid wow region"),
			},
		}
	}

	exists, err := api_utils.CheckWowCharacterExists(&character)
	if err != nil {
		return fmt.Errorf("check character existence: %w", err)
	}

	if !exists {
		log.Printf("WoW character does not exist")

		return &wowCharacterValidationError{
			statusCode: http.StatusNotFound,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("WoW character does not exist"),
			},
		}
	}

	return nil
}
