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

	"github.com/DomNidy/saint_sim/apps/api/middleware"
	"github.com/DomNidy/saint_sim/pkg/api_types"
	"github.com/DomNidy/saint_sim/pkg/db"
	"github.com/DomNidy/saint_sim/pkg/utils"
)

type simulationValidationError struct {
	statusCode int
	response   api_types.ErrorResponse
}

func (failure *simulationValidationError) Error() string {
	if failure == nil || failure.response.Message == nil {
		return "simulation validation failed"
	}

	return *failure.response.Message
}

type simulationQueue interface {
	Publish(simJob utils.SimulationJobMessage) error
}

type simulationStore interface {
	CreateSimulation(ctx context.Context, arg db.CreateSimulationParams) (db.Simulation, error)
}

// Simulate validates a simulation request, persists it, and enqueues it for processing.
func Simulate(
	ginContext *gin.Context,
	dbClient simulationStore,
	simQueue simulationQueue,
) {
	var simOptions api_types.SimulationOptions

	err := ginContext.ShouldBindJSON(&simOptions)
	if err != nil {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("Invalid simulation options"),
		})

		return
	}

	errValidate := validateSimulationRequest(ginContext, simOptions)
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

	SimulationJobMessageBody := utils.SimulationJobMessage{
		SimulationID: simulationID,
	}

	err = simQueue.Publish(SimulationJobMessageBody)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)
		ginContext.JSON(http.StatusInternalServerError, api_types.ErrorResponse{
			Message: utils.StrPtr("An internal server error occurred. Please try again later."),
		})

		return
	}

	log.Printf(" [x] Sent %v\n", SimulationJobMessageBody)
	ginContext.JSON(http.StatusAccepted, gin.H{
		"simulation_id": SimulationJobMessageBody.SimulationID,
	})
}

func validateSimulationRequest(
	ctx context.Context,
	simOptions api_types.SimulationOptions,
) *simulationValidationError {
	_ = ctx

	err := utils.ValidateSimOptions(&simOptions)
	if err != nil {
		return &simulationValidationError{
			statusCode: http.StatusBadRequest,
			response: api_types.ErrorResponse{
				Message: utils.StrPtr("Bad request"),
			},
		}
	}

	return nil
}

func createSimulationRequest(
	ginContext *gin.Context,
	dbClient simulationStore,
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

func simulationOwnerID(ginContext *gin.Context) *string {
	userID, ok := middleware.GetEffectiveUserID(ginContext)
	if !ok {
		return nil
	}

	return &userID
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

	response := simulationResponseFromRecord(simulation)

	ginContext.JSON(http.StatusOK, response)
}

func simulationResponseFromRecord(simulation db.Simulation) api_types.Simulation {
	var response api_types.Simulation

	status := simulationStatusFromRecord(simulation)

	simulationID := simulation.ID.String()
	response.Id = &simulationID
	response.SimulationStatus = &status

	if simulation.SimResult != nil {
		response.SimResult = simulation.SimResult
	}

	if simulation.ErrorText != nil {
		response.ErrorText = simulation.ErrorText
	}

	return response
}

func simulationStatusFromRecord(simulation db.Simulation) api_types.SimulationStatus {
	if simulation.ErrorText != nil {
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
