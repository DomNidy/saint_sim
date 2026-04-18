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
	"github.com/DomNidy/saint_sim/internal/simc"
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
			Message: utils.StrPtr("Invalid simulation options"),
		}, nil
	}

	authContext, ok := auth.ResolveAuthFromContext(ctx)
	if !ok {
		return api.Simulate401JSONResponse{
			Message: utils.StrPtr("Unauthorized"),
		}, nil
	}

	simOptions := *request.Body

	options, err := simOptions.ValueByDiscriminator()
	if err != nil {
		return api.Simulate422JSONResponse{
			MalformedRequestJSONResponse: api.MalformedRequestJSONResponse{
				Message: "Invalid or malformed input",
			},
		}, nil
	}

	if topGearOptions, ok := options.(api.SimulationOptionsTopGear); ok {
		log.Printf("got a topgear sim, but not implemented yet: %v", topGearOptions)

		return api.Simulate202JSONResponse{
			SimulationId: utils.StrPtr("not_implemented_yet_id_123"),
		}, nil
	}

	if basicSimOptions, ok := options.(api.SimulationOptionsBasic); ok {
		return server.handleSimulationOptionsBasic(ctx, authContext, basicSimOptions)
	}

	// if we reach here, that means we didn't handle all possible simulation option types
	// so 500 error
	return api.Simulate500JSONResponse{
		InternalErrorJSONResponse: api.InternalErrorJSONResponse{
			Message: utils.StrPtr("Internal server error try again later"),
		},
	}, nil
}

func (server *Server) handleSimulationOptionsBasic(
	ctx context.Context,
	authContext auth.AuthContext,
	simOptions api.SimulationOptionsBasic,
) (api.SimulateResponseObject, error) {
	simOptions.SimcAddonExport = simc.NormalizeLineEndings(simOptions.SimcAddonExport)

	validationFailure := validateSimulationRequestBasic(ctx, simOptions)
	if validationFailure != nil {
		return api.Simulate400JSONResponse(validationFailure.response), nil
	}

	simulationID, err := createSimulationRequestBasic(ctx, authContext, server.dbClient, simOptions)
	if err != nil {
		log.Printf("Error creating simulation request: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: utils.StrPtr("Internal server error"),
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
				Message: utils.StrPtr("An internal server error occurred. Please try again later."),
			},
		}, nil
	}

	log.Printf(" [x] Sent %v\n", simulationJobMessage)

	return api.Simulate202JSONResponse{
		SimulationId: &simulationID,
	}, nil
}

func validateSimulationRequestBasic(
	ctx context.Context,
	simOptions api.SimulationOptionsBasic,
) *simulationValidationError {
	_ = ctx

	err := utils.ValidateSimulationOptionsBasic(&simOptions)
	if err != nil {
		return &simulationValidationError{
			statusCode: http.StatusBadRequest,
			response: api.ErrorResponse{
				Message: utils.StrPtr("Bad request"),
			},
		}
	}

	return nil
}

func createSimulationRequestBasic(
	ctx context.Context,
	authContext auth.AuthContext,
	dbClient simulationCreator,
	simOptions api.SimulationOptionsBasic,
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
