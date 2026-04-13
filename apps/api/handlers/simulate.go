// Package handlers contains the Gin handlers for the API server.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/DomNidy/saint_sim/apps/api/middleware"
	"github.com/DomNidy/saint_sim/pkg/api_types"
	"github.com/DomNidy/saint_sim/pkg/db"
	"github.com/DomNidy/saint_sim/pkg/utils"
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
	httpClient *http.Client,
) {
	var simOptions api_types.SimulationOptions

	err := ginContext.ShouldBindJSON(&simOptions)
	if err != nil {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("Invalid simulation options"),
		})

		return
	}

	errValidate := validateSimulationRequest(ginContext, httpClient, simOptions)
	if errValidate != nil {
		ginContext.JSON(errValidate.statusCode, errValidate.response)

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

func validateSimulationRequest(
	ctx context.Context,
	httpClient *http.Client,
	simOptions api_types.SimulationOptions,
) *wowCharacterValidationError {
	if !utils.IsValidSimOptions(&simOptions) {
		return &wowCharacterValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Bad request"),
			},
		}
	}

	err := api_utils.CheckWowCharacterExists(ctx, httpClient, &simOptions.WowCharacter)
	if err != nil {
		log.Printf("%v", err)

		if errors.Is(err, api_utils.ErrCharacterNotExistsOnArmory) {
			return &wowCharacterValidationError{
				statusCode: http.StatusNotFound,
				response: api_types.ErrorResponse{
					Message: utils.StrPtr("Character not found"),
				},
			}
		}

		if errors.Is(err, api_utils.ErrUnexpectedStatusCodeReceivedFromArmory) {
			return &wowCharacterValidationError{
				statusCode: http.StatusInternalServerError,
				response: api_types.ErrorResponse{
					Message: utils.StrPtr("Internal server error"),
				},
			}
		}
	}

	return nil
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

	simEntry, err := dbClient.CreateSimulation(
		ginContext.Request.Context(),
		db.CreateSimulationParams{
			SimConfig: simOptionsJSON,
			OwnerID:   simulationOwnerID(ginContext),
		},
	)
	if err != nil {
		return "", fmt.Errorf("create simulation row: %w", err)
	}

	return simEntry.ID.String(), nil
}

func simulationOwnerID(ginContext *gin.Context) pgtype.Text {
	userID, ok := middleware.GetEffectiveUserID(ginContext)
	if !ok {
		return pgtype.Text{
			String: "",
			Valid:  false,
		}
	}

	return pgtype.Text{
		String: userID,
		Valid:  true,
	}
}

// GetSimulation returns the current state or completed result for a simulation.
func GetSimulation(ginContext *gin.Context, dbClient *db.Queries) {
	simulationID, err := uuid.Parse(ginContext.Param("id"))
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
	var response api_types.Simulation

	status := simulationStatusFromRecord(simulation)

	simulationID := simulation.ID.String()
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
