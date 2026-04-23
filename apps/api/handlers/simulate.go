package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	"github.com/DomNidy/saint_sim/apps/api/usecases"
	api "github.com/DomNidy/saint_sim/internal/api"
)

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

	simulationID, err := server.submitSimulation.Submit(
		ctx,
		usecases.SubmitSimulationInput{
			Options: *request.Body,
			OwnerID: simulationOwnerID(authContext),
		},
	)
	if err == nil {
		simulationIDString := simulationID.String()

		return api.Simulate202JSONResponse{
			SimulationId: &simulationIDString,
		}, nil
	}

	switch {
	case errors.Is(err, usecases.ErrMalformedSimulationInput):
		return api.Simulate422JSONResponse{
			MalformedRequestJSONResponse: api.MalformedRequestJSONResponse{
				Message: "Invalid or malformed input",
				Code:    "",
			},
		}, nil
	case errors.Is(err, usecases.ErrInvalidSimulationInput):
		return api.Simulate400JSONResponse{
			Message: "Bad request",
			Code:    "",
		}, nil
	case errors.Is(err, usecases.ErrUnsupportedSimulationInput):
		log.Printf("unsupported simulation input: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "Internal server error try again later",
				Code:    "",
			},
		}, nil
	default:
		log.Printf("Error submitting simulation request: %v", err)

		return api.Simulate500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: "Internal server error",
				Code:    "",
			},
		}, nil
	}
}

func simulationOwnerID(authContext auth.AuthContext) *string {
	userID, ok := auth.EffectiveUserID(authContext)
	if !ok {
		return nil
	}

	return &userID
}
