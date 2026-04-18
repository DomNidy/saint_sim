// Package handlers contains the OpenAPI server implementation for the API.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

type simulationQueue interface {
	Publish(simJob utils.SimulationJobMessage) error
}

type simulationCreator interface {
	CreateSimulation(ctx context.Context, arg db.CreateSimulationParams) (db.Simulation, error)
}

type simulationReader interface {
	GetSimulation(ctx context.Context, id uuid.UUID) (db.Simulation, error)
}

type simulationRepository interface {
	simulationCreator
	simulationReader
}

// Server implements the generated strict OpenAPI server.
type Server struct {
	dbClient simulationRepository
	simQueue simulationQueue
}

// NewServer constructs the strict OpenAPI server implementation.
func NewServer(dbClient simulationRepository, simQueue simulationQueue) *Server {
	return &Server{
		dbClient: dbClient,
		simQueue: simQueue,
	}
}

// Health returns the API health status.
func (server *Server) Health(
	_ context.Context,
	request api.HealthRequestObject,
) (api.HealthResponseObject, error) {
	_ = server
	_ = request

	return api.Health200JSONResponse{
		Status: "healthy",
	}, nil
}

// ParseAddonExport parses a SimC addon export and returns structured data.
func (server *Server) ParseAddonExport(
	_ context.Context,
	request api.ParseAddonExportRequestObject,
) (api.ParseAddonExportResponseObject, error) {
	_ = server

	if request.Body == nil {
		return api.ParseAddonExport400JSONResponse{
			BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
				Message: utils.StrPtr("Invalid parse addon export request"),
			},
		}, nil
	}

	normalizedExport := simc.NormalizeLineEndings(request.Body.SimcAddonExport)
	if strings.TrimSpace(normalizedExport) == "" {
		return api.ParseAddonExport400JSONResponse{
			BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
				Message: utils.StrPtr("simc_addon_export is required"),
			},
		}, nil
	}

	addonExport := simc.Parse(normalizedExport)
	if !simc.HasRecognizedData(addonExport) {
		return api.ParseAddonExport400JSONResponse{
			BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
				Message: utils.StrPtr("no recognizable addon export data found"),
			},
		}, nil
	}

	return api.ParseAddonExport200JSONResponse{
		AddonExport: addonExport,
	}, nil
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
