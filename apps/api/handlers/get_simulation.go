package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
)

// GetSimulation returns the current state of a simulation job. The response
// is the polling envelope: status + kind + (when complete) the typed result
// as written by the worker. The API never recomputes the result — it
// returns `simulation.sim_result` verbatim, rehydrated from jsonb into the
// discriminated union.
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

	response, err := simulationResponseFromRecord(simulation)
	if err != nil {
		log.Printf("Error building simulation response %s: %v", request.Id.String(), err)

		return internalErrorResponse(), nil
	}

	return api.GetSimulation200JSONResponse(response), nil
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
	var empty db.Simulation

	simulation, err := dbClient.GetSimulation(ctx, simulationID)
	if err == nil {
		return simulation, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return empty, simulationNotFoundResponse()
	}

	log.Printf("Error loading simulation %s: %v", rawSimulationID, err)
	return empty, internalErrorResponse()
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

// simulationResponseFromRecord projects a simulation row into the API
// envelope. When sim_result is populated (status == complete) it's
// rehydrated into the SimulationResult union via its generated
// UnmarshalJSON, which just stores the bytes — the client does the actual
// discriminated‑union decoding on the other side.
func simulationResponseFromRecord(s db.Simulation) (api.Simulation, error) {
	response := api.Simulation{
		Id:        s.ID,
		Kind:      api.SimulationKind(s.Kind),
		Status:    api.SimulationStatus(s.Status),
		ErrorText: s.ErrorText,
	}

	if len(s.SimResult) > 0 {
		var result api.SimulationResult
		if err := result.UnmarshalJSON(s.SimResult); err != nil {
			return api.Simulation{}, err
		}
		response.Result = &result
	}

	return response, nil
}
