package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	api_types "github.com/DomNidy/saint_sim/pkg/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/db"
)

type simulationStore struct {
	queries dbqueries.Queries
}

type simulationRequest struct {
	id      uuid.UUID
	idText  string
	options api_types.SimulationOptions
}

func (store simulationStore) LoadRequest(
	ctx context.Context,
	requestID uuid.UUID,
	requestIDText string,
) (simulationRequest, error) {
	simOptionsJSON, err := store.queries.GetSimulationOptions(ctx, requestID)
	if err != nil {
		return simulationRequest{}, fmt.Errorf("retrieve simulation options from db: %w", err)
	}

	var options api_types.SimulationOptions

	err = json.Unmarshal(simOptionsJSON, &options)
	if err != nil {
		return simulationRequest{}, fmt.Errorf("unmarshal simulation options: %w", err)
	}

	return simulationRequest{
		id:      requestID,
		idText:  requestIDText,
		options: options,
	}, nil
}

func (store simulationStore) MarkStarted(ctx context.Context, requestID uuid.UUID) error {
	_, err := store.queries.UpdateSimulation(
		ctx,
		dbqueries.UpdateSimulationParams{
			SimResult:   invalidText(),
			ErrorText:   invalidText(),
			StartedAt:   timestampValue(time.Now()),
			CompletedAt: invalidTimestamp(),
			ID:          requestID,
		},
	)
	if err != nil {
		return fmt.Errorf("mark simulation started in db: %w", err)
	}

	return nil
}

func (store simulationStore) MarkCompleted(
	ctx context.Context,
	requestID uuid.UUID,
	simulationResult []byte,
) error {
	simResult := pgtype.Text{
		String: string(simulationResult),
		Valid:  true,
	}

	_, err := store.queries.UpdateSimulation(
		ctx,
		dbqueries.UpdateSimulationParams{
			SimResult:   simResult,
			ErrorText:   invalidText(),
			StartedAt:   invalidTimestamp(),
			CompletedAt: timestampValue(time.Now()),
			ID:          requestID,
		},
	)
	if err != nil {
		return fmt.Errorf("write simulation result to db: %w", err)
	}

	return nil
}

func (store simulationStore) MarkFailed(
	ctx context.Context,
	requestID uuid.UUID,
	cause error,
) error {
	_, err := store.queries.UpdateSimulation(
		ctx,
		dbqueries.UpdateSimulationParams{
			SimResult: invalidText(),
			ErrorText: pgtype.Text{
				String: cause.Error(),
				Valid:  true,
			},
			StartedAt:   invalidTimestamp(),
			CompletedAt: timestampValue(time.Now()),
			ID:          requestID,
		},
	)
	if err != nil {
		return fmt.Errorf("mark simulation failed in db: %w", err)
	}

	return nil
}

func timestampValue(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             value,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func invalidText() pgtype.Text {
	return pgtype.Text{
		String: "",
		Valid:  false,
	}
}

func invalidTimestamp() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time.Time{},
		InfinityModifier: pgtype.Finite,
		Valid:            false,
	}
}
