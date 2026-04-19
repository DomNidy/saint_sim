package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	api "github.com/DomNidy/saint_sim/internal/api"
	dbqueries "github.com/DomNidy/saint_sim/internal/db"
)

type Store interface {
	LoadRequest(ctx context.Context, requestID uuid.UUID) (simulationRequest, error)
	UpdateSimulation(
		ctx context.Context,
		updateOptions dbqueries.UpdateSimulationParams,
	) (dbqueries.Simulation, error)
}

type dbSimulationStore struct {
	queries dbqueries.Queries
}

type simulationRequest struct {
	id      uuid.UUID
	options api.SimulationOptions
}

func (store dbSimulationStore) LoadRequest(
	ctx context.Context,
	requestID uuid.UUID,
) (simulationRequest, error) {
	simOptionsJSON, err := store.queries.GetSimulationOptions(ctx, requestID)
	if err != nil {
		return simulationRequest{}, fmt.Errorf("retrieve simulation options from db: %w", err)
	}

	var options api.SimulationOptions

	err = json.Unmarshal(simOptionsJSON, &options)
	if err != nil {
		return simulationRequest{}, fmt.Errorf("unmarshal simulation options: %w", err)
	}

	return simulationRequest{
		id:      requestID,
		options: options,
	}, nil
}

func (store dbSimulationStore) UpdateSimulation(
	ctx context.Context,
	updateOptions dbqueries.UpdateSimulationParams,
) (dbqueries.Simulation, error) {
	return store.queries.UpdateSimulation(ctx, updateOptions)
}

func timestampValue(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             value,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func invalidTimestamp() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time.Time{},
		InfinityModifier: pgtype.Finite,
		Valid:            false,
	}
}
