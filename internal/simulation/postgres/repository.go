package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DomNidy/saint_sim/internal/api"
	dbqueries "github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

const defaultSimulationErrorText = "internal server error"

// Repository adapts generated sqlc queries to simulation use-case repositories.
type Repository struct {
	queries *dbqueries.Queries
}

// NewRepository constructs a Postgres-backed simulation repository.
func NewRepository(db dbqueries.DBTX) *Repository {
	return &Repository{
		queries: dbqueries.New(db),
	}
}

// CreateQueuedSimulation persists a queued simulation request.
func (repo *Repository) CreateQueuedSimulation(
	ctx context.Context,
	input simulation.CreateQueuedSimulationInput,
) (uuid.UUID, error) {
	simOptionsJSON, err := json.Marshal(input.Options)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("marshal simulation options: %w", err)
	}

	simEntry, err := repo.queries.CreateSimulation(ctx, dbqueries.CreateSimulationParams{
		Kind:      dbqueries.SimulationKind(input.Kind),
		SimConfig: simOptionsJSON,
		OwnerID:   input.OwnerID,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create simulation row: %w", err)
	}

	return simEntry.ID, nil
}

// GetSimulation returns the API polling envelope for a simulation.
func (repo *Repository) GetSimulation(ctx context.Context, id uuid.UUID) (api.Simulation, error) {
	simRecord, err := repo.queries.GetSimulation(ctx, id)
	if err != nil {
		return api.Simulation{}, translateNotFound(err)
	}

	response := api.Simulation{
		Id:           simRecord.ID,
		Kind:         api.SimulationKind(simRecord.Kind),
		Status:       api.SimulationStatus(simRecord.Status),
		ErrorText:    simRecord.ErrorText,
		RawSimcInput: simRecord.RawSimcInput,
	}

	if len(simRecord.SimResult) > 0 {
		var result api.SimulationResult
		if err := result.UnmarshalJSON(simRecord.SimResult); err != nil {
			return api.Simulation{}, fmt.Errorf("unmarshal simulation result: %w", err)
		}
		response.Result = &result
	}

	return response, nil
}

// LoadRequest loads and decodes the simulation config for the worker.
func (repo *Repository) LoadRequest(
	ctx context.Context,
	requestID uuid.UUID,
) (simulation.SimulationRequest, error) {
	simOptionsJSON, err := repo.queries.GetSimulationOptions(ctx, requestID)
	if err != nil {
		return simulation.SimulationRequest{}, translateNotFound(err)
	}

	var options api.SimulationOptions
	if err := json.Unmarshal(simOptionsJSON, &options); err != nil {
		return simulation.SimulationRequest{}, fmt.Errorf("unmarshal simulation options: %w", err)
	}

	return simulation.SimulationRequest{
		ID:      requestID,
		Options: options,
	}, nil
}

// MarkInProgress records that the worker started processing a simulation.
func (repo *Repository) MarkInProgress(ctx context.Context, id uuid.UUID) error {
	_, err := repo.queries.UpdateSimulation(ctx, dbqueries.UpdateSimulationParams{
		ID:        id,
		StartedAt: nowTimestamptz(),
		Status: dbqueries.NullSimulationStatus{
			SimulationStatus: dbqueries.SimulationStatusInProgress,
			Valid:            true,
		},
	})
	return translateNotFound(err)
}

// MarkCompleted records the user-facing result and raw SimC artifact.
func (repo *Repository) MarkCompleted(
	ctx context.Context,
	id uuid.UUID,
	result simulation.CompletedSimulation,
) error {
	var json2Bytes []byte
	var err error
	if result.RawJSON2 != nil {
		json2Bytes, err = result.RawJSON2.Marshal()
		if err != nil {
			log.Printf(
				"WARN: failed to marshal rawJson2 into a byte array - db will be missing raw JSON2 output for this sim: %v",
				err,
			)
		}
	}

	simResultBytes, err := json.Marshal(result.Result)
	if err != nil {
		return fmt.Errorf("marshal simulation result: %w", err)
	}

	_, err = repo.queries.UpdateSimulation(ctx, dbqueries.UpdateSimulationParams{
		ID:           id,
		SimResult:    simResultBytes,
		SimcRawJson2: json2Bytes,
		CompletedAt:  nowTimestamptz(),
		Status: dbqueries.NullSimulationStatus{
			SimulationStatus: dbqueries.SimulationStatusComplete,
			Valid:            true,
		},
	})
	return translateNotFound(err)
}

// MarkFailed records a terminal simulation failure using sanitized text.
func (repo *Repository) MarkFailed(
	ctx context.Context,
	id uuid.UUID,
	failure simulation.FailedSimulation,
) error {
	errorText := failure.ErrorText
	if errorText == "" {
		errorText = defaultSimulationErrorText
	}

	_, err := repo.queries.UpdateSimulation(ctx, dbqueries.UpdateSimulationParams{
		ID:        id,
		ErrorText: &errorText,
		Status: dbqueries.NullSimulationStatus{
			SimulationStatus: dbqueries.SimulationStatusError,
			Valid:            true,
		},
	})

	return translateNotFound(err)
}

func (repo *Repository) WriteRunDetails(
	ctx context.Context,
	requestID uuid.UUID,
	rawProfileText string,
) error {
	_, err := repo.queries.UpdateSimulation(ctx, dbqueries.UpdateSimulationParams{
		ID:           requestID,
		RawSimcInput: &rawProfileText,
	})

	return translateNotFound(err)
}

func translateNotFound(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return simulation.ErrNotFound
	}

	return err
}

func nowTimestamptz() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time.Now().UTC(),
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}
