package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

// GetSimulation returns the current state or completed result for a simulation.
func (server *Server) GetSimulation(
	ctx context.Context,
	request api.GetSimulationRequestObject,
) (api.GetSimulationResponseObject, error) {
	simulationID, validID := parseSimulationID(request.Id.String())
	if !validID {
		return simulationNotFoundResponse(), nil
	}

	simulation, errorResponse := loadSimulation(
		ctx,
		server.dbClient,
		request.Id.String(),
		simulationID,
	)
	if errorResponse != nil {
		return errorResponse, nil
	}

	return api.GetSimulation200JSONResponse(simulationResponseFromRecord(simulation)), nil
}

func parseSimulationID(rawSimulationID string) (uuid.UUID, bool) {
	simulationID, err := uuid.Parse(rawSimulationID)
	if err != nil {
		return uuid.UUID{}, false
	}

	return simulationID, true
}

func loadSimulation(
	ctx context.Context,
	dbClient simulationReader,
	rawSimulationID string,
	simulationID uuid.UUID,
) (db.Simulation, api.GetSimulationResponseObject) {
	var emptySimulation db.Simulation

	simulation, err := dbClient.GetSimulation(ctx, simulationID)
	if err == nil {
		return simulation, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return emptySimulation, simulationNotFoundResponse()
	}

	log.Printf("Error loading simulation %s: %v", rawSimulationID, err)

	return emptySimulation, api.GetSimulation500JSONResponse{
		InternalErrorJSONResponse: api.InternalErrorJSONResponse{
			Message: utils.StrPtr("Internal server error"),
		},
	}
}

func simulationNotFoundResponse() api.GetSimulation404JSONResponse {
	return api.GetSimulation404JSONResponse{
		NotFoundErrorJSONResponse: api.NotFoundErrorJSONResponse{
			Message: utils.StrPtr("Simulation not found"),
		},
	}
}

func simulationResponseFromRecord(simulation db.Simulation) api.Simulation {
	var response api.Simulation

	status := simulationStatusFromRecord(simulation)

	response.Kind = api.SimulationKind(simulation.Kind)
	response.Id = simulation.ID
	response.Status = status
	// response.Result = simulation.SimResult

	if simulation.ErrorText != nil {
		response.ErrorText = simulation.ErrorText
	}

	return response
}

func simulationStatusFromRecord(simulation db.Simulation) api.SimulationStatus {
	if simulation.ErrorText != nil {
		return api.Error
	}

	if simulation.CompletedAt.Valid {
		return api.Complete
	}

	if simulation.StartedAt.Valid {
		return api.InProgress
	}

	return api.InQueue
}
