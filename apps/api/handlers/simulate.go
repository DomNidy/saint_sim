package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

type simulationValidationError struct {
	statusCode int
	response   api.ErrorResponse
}

// Simulate validates a simulation request, persists it, and enqueues it for processing.
func (server *Server) Simulate(
	ctx context.Context,
	request api.SimulateRequestObject,
) (api.SimulateResponseObject, error) {
	if request.Body == nil {
		return api.Simulate400JSONResponse{
			Message: "Invalid simulation options",
			Code:    "",
		}, nil
	}

	authContext, authResolved := auth.ResolveAuthFromContext(ctx)
	if !authResolved {
		log.Printf("simulate unauthorized: auth context missing from request context")
		return api.Simulate401JSONResponse{
			Message: "Unauthorized",
			Code:    "",
		}, nil
	}

	simOptions := *request.Body

	config, ok := simulationConfigByDiscriminator(simOptions)
	if !ok {
		return api.Simulate422JSONResponse{
			MalformedRequestJSONResponse: api.MalformedRequestJSONResponse{
				Message: "Invalid or malformed input",
			},
		}, nil
	}

	if topGearConfig, ok := config.(api.SimulationConfigTopGear); ok {
		return server.handleSimulationTopGear(ctx, authContext, topGearConfig)
	}

	if basicSimConfig, ok := config.(api.SimulationConfigBasic); ok {
		return server.handleSimulationBasic(ctx, authContext, basicSimConfig)
	}

	// if we reach here, that means we didn't handle all possible simulation option types
	// so 500 error
	return api.Simulate500JSONResponse{
		InternalErrorJSONResponse: api.InternalErrorJSONResponse{
			Message: "Internal server error try again later",
			Code:    "",
		},
	}, nil
}

func (server *Server) handleSimulationBasic(
	ctx context.Context,
	authContext auth.AuthContext,
	simConfig api.SimulationConfigBasic,
) (api.SimulateResponseObject, error) {
	validationFailure := validateSimulationRequestBasic(ctx, simConfig)
	if validationFailure != nil {
		return api.Simulate400JSONResponse(validationFailure.response), nil
	}

	simulationID, err := createSimulationRequest(
		ctx,
		authContext,
		server.dbClient,
		db.SimulationKindBasic,
		simConfig,
	)
	if err != nil {
		log.Printf("Error creating simulation request: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "Internal server error",
				Code:    "",
			},
		}, nil
	}

	simulationJobMessage := utils.SimulationJobMessage{
		SimulationID: simulationID,
	}

	err = server.simQueue.Publish(simulationJobMessage)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "An internal server error occurred. Please try again later.",
				Code:    "",
			},
		}, nil
	}

	log.Printf(" [x] Sent %v\n", simulationJobMessage)

	return api.Simulate202JSONResponse{
		SimulationId: &simulationID,
	}, nil
}

func (server *Server) handleSimulationTopGear(
	ctx context.Context,
	authContext auth.AuthContext,
	simOptions api.SimulationConfigTopGear,
) (api.SimulateResponseObject, error) {
	simulationID, err := createSimulationRequest(
		ctx,
		authContext,
		server.dbClient,
		db.SimulationKindTopGear,
		simOptions,
	)
	if err != nil {
		log.Printf("Error creating top gear simulation request: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "Internal server error",
				Code:    "",
			},
		}, nil
	}

	simulationJobMessage := utils.SimulationJobMessage{
		SimulationID: simulationID,
	}

	err = server.simQueue.Publish(simulationJobMessage)
	if err != nil {
		log.Printf("ERROR: Failed to publish simulation message to queue: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "An internal server error occurred. Please try again later.",
				Code:    "",
			},
		}, nil
	}

	log.Printf(" [x] Sent %v\n", simulationJobMessage)

	return api.Simulate202JSONResponse{
		SimulationId: &simulationID,
	}, nil
}

func simulationConfigByDiscriminator(config api.SimulationOptions) (interface{}, bool) {
	simConfig, err := config.ValueByDiscriminator()
	if err != nil {
		return nil, false
	}

	return simConfig, true
}

func validateSimulationRequestBasic(
	ctx context.Context,
	simOptions api.SimulationConfigBasic,
) *simulationValidationError {
	_ = ctx

	err := utils.ValidateSimulationConfigBasic(&simOptions)
	if err != nil {
		return &simulationValidationError{
			statusCode: http.StatusBadRequest,
			response: api.ErrorResponse{
				Message: "Bad request",
				Code:    "",
			},
		}
	}

	return nil
}

func createSimulationRequest[T any](
	ctx context.Context,
	authContext auth.AuthContext,
	dbClient simulationCreator,
	kind db.SimulationKind,
	simOptions T,
) (string, error) {
	simOptionsJSON, err := json.Marshal(simOptions)
	if err != nil {
		return "", fmt.Errorf("marshal simulation options: %w", err)
	}

	simEntry, err := dbClient.CreateSimulation(
		ctx,
		db.CreateSimulationParams{
			SimConfig: simOptionsJSON,
			OwnerID:   simulationOwnerID(authContext),
			Kind:      kind,
		},
	)
	if err != nil {
		return "", fmt.Errorf("create simulation row: %w", err)
	}

	return simEntry.ID.String(), nil
}

func simulationOwnerID(authContext auth.AuthContext) *string {
	userID, ok := auth.EffectiveUserID(authContext)
	if !ok {
		return nil
	}

	return &userID
}
