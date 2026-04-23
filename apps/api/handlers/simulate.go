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

type simulationValidator[T any] func(context.Context, T) *simulationValidationError

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
				Code:    "",
			},
		}, nil
	}

	switch simConfig := config.(type) {
	case api.SimulationConfigBasic:
		return submitSimulation(
			ctx,
			server,
			authContext,
			db.SimulationKindBasic,
			simConfig,
			validateSimulationConfigBasic,
		)
	case api.SimulationConfigTopGear:
		return submitSimulation(
			ctx,
			server,
			authContext,
			db.SimulationKindTopGear,
			simConfig,
			noValidation[api.SimulationConfigTopGear],
		)
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

func noValidation[T any](context.Context, T) *simulationValidationError {
	return nil
}

func submitSimulation[T api.SimulationConfigBasic | api.SimulationConfigTopGear](
	ctx context.Context,
	server *Server,
	authContext auth.AuthContext,
	kind db.SimulationKind,
	simConfig T,
	validate simulationValidator[T],
) (api.SimulateResponseObject, error) {
	if validationFailure := validate(ctx, simConfig); validationFailure != nil {
		return api.Simulate400JSONResponse(validationFailure.response), nil
	}

	simulationID, err := createSimulationRequest(
		ctx,
		authContext,
		server.dbClient,
		kind,
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

func simulationConfigByDiscriminator(config api.SimulationOptions) (interface{}, bool) {
	simConfig, err := config.ValueByDiscriminator()
	if err != nil {
		return nil, false
	}

	return simConfig, true
}

func validateSimulationConfigBasic(
	ctx context.Context,
	simConfig api.SimulationConfigBasic,
) *simulationValidationError {
	_ = ctx

	err := utils.ValidateSimulationConfigBasic(&simConfig)
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

func createSimulationRequest[T api.SimulationConfigBasic | api.SimulationConfigTopGear](
	ctx context.Context,
	authContext auth.AuthContext,
	dbClient simulationCreator,
	kind db.SimulationKind,
	simConfig T,
) (string, error) {
	simOptionsJSON, err := json.Marshal(simConfig)
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
