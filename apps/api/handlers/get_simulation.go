package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

// GetSimulation returns the current state of a simulation job.
func (server *Server) GetSimulation(
	ctx context.Context,
	request api.GetSimulationRequestObject,
) (api.GetSimulationResponseObject, error) {
	simulationID, validID := parseSimulationID(request.Id.String())
	if !validID {
		return simulationNotFoundResponse(), nil
	}

	response, err := server.getSimulation.Get(ctx, simulationID)
	if err == nil {
		return api.GetSimulation200JSONResponse(response), nil
	}

	if errors.Is(err, simulation.ErrNotFound) {
		return simulationNotFoundResponse(), nil
	}

	log.Printf("Error loading simulation %s: %v", request.Id.String(), err)

	return internalErrorResponse(), nil
}

func parseSimulationID(rawSimulationID string) (uuid.UUID, bool) {
	simulationID, err := uuid.Parse(rawSimulationID)
	if err != nil {
		return uuid.UUID{}, false
	}

	return simulationID, true
}

func simulationNotFoundResponse() api.GetSimulation404JSONResponse {
	return api.GetSimulation404JSONResponse{
		NotFoundErrorJSONResponse: api.NotFoundErrorJSONResponse{
			Message: "Simulation not found",
			Code:    "",
		},
	}
}

func internalErrorResponse() api.GetSimulation500JSONResponse {
	return api.GetSimulation500JSONResponse{
		InternalErrorJSONResponse: api.InternalErrorJSONResponse{
			Message: "Internal server error",
			Code:    "",
		},
	}
}
