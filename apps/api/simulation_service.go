package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidSimOptions = errors.New("invalid simulation options")
	ErrInvalidWowRealm   = errors.New("invalid wow realm")
	ErrInvalidWowRegion  = errors.New("invalid wow region")
	ErrCharacterNotFound = errors.New("wow character does not exist")
)

type SimulationRequestStore interface {
	CreateSimulationRequest(ctx context.Context, arg dbqueries.CreateSimulationRequestParams) error
}

type SimulationDispatcher interface {
	DispatchSimulation(ctx context.Context, msg api_types.SimulationMessageBody) error
}

type CharacterLookup interface {
	Exists(character *api_types.WowCharacter) (bool, error)
}

type SimulationService struct {
	store      SimulationRequestStore
	dispatcher SimulationDispatcher
	characters CharacterLookup
	idgen      func() string
}

func NewSimulationService(store SimulationRequestStore, dispatcher SimulationDispatcher, characters CharacterLookup, idgen func() string) SimulationService {
	return SimulationService{
		store:      store,
		dispatcher: dispatcher,
		characters: characters,
		idgen:      idgen,
	}
}

func (s SimulationService) Submit(ctx context.Context, simOptions api_types.SimulationOptions) (*api_types.SimulationResponse, error) {
	if !utils.IsValidSimOptions(&simOptions) {
		return nil, ErrInvalidSimOptions
	}

	if !utils.IsValidWowRealm(string(simOptions.WowCharacter.Realm)) {
		return nil, ErrInvalidWowRealm
	}

	if !utils.IsValidWowRegion(string(simOptions.WowCharacter.Region)) {
		return nil, ErrInvalidWowRegion
	}

	exists, err := s.characters.Exists(&simOptions.WowCharacter)
	if err != nil {
		return nil, fmt.Errorf("check wow character exists: %w", err)
	}
	if !exists {
		return nil, ErrCharacterNotFound
	}

	receivedJSON, err := json.Marshal(simOptions)
	if err != nil {
		return nil, fmt.Errorf("marshal simulation options: %w", err)
	}

	simulationRequestID := s.idgen()

	var simulationRequestUUID pgtype.UUID
	if err := simulationRequestUUID.Scan(simulationRequestID); err != nil {
		return nil, fmt.Errorf("convert simulation request id to uuid: %w", err)
	}

	if err := s.store.CreateSimulationRequest(ctx, dbqueries.CreateSimulationRequestParams{
		ID:      simulationRequestUUID,
		Options: receivedJSON,
	}); err != nil {
		return nil, fmt.Errorf("create simulation request: %w", err)
	}

	if err := s.dispatcher.DispatchSimulation(ctx, api_types.SimulationMessageBody{
		SimulationId: &simulationRequestID,
	}); err != nil {
		return nil, fmt.Errorf("dispatch simulation request: %w", err)
	}

	return &api_types.SimulationResponse{
		SimulationRequestId: &simulationRequestID,
	}, nil
}
