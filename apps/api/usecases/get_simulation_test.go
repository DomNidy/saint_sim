package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

type stubGetRepository struct {
	get func(context.Context, uuid.UUID) (api.Simulation, error)
}

func (repo stubGetRepository) GetSimulation(
	ctx context.Context,
	id uuid.UUID,
) (api.Simulation, error) {
	return repo.get(ctx, id)
}

func TestGetSimulationReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	simulationID := uuid.New()
	want := api.Simulation{
		Id:     simulationID,
		Kind:   api.SimulationKindBasic,
		Status: api.Complete,
	}
	useCase := NewGetSimulationUseCase(stubGetRepository{
		get: func(_ context.Context, id uuid.UUID) (api.Simulation, error) {
			if id != simulationID {
				t.Fatalf("id = %s, want %s", id, simulationID)
			}

			return want, nil
		},
	})

	got, err := useCase.Get(t.Context(), simulationID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != want {
		t.Fatalf("Get() = %#v, want %#v", got, want)
	}
}

func TestGetSimulationPropagatesNotFound(t *testing.T) {
	t.Parallel()

	useCase := NewGetSimulationUseCase(stubGetRepository{
		get: func(context.Context, uuid.UUID) (api.Simulation, error) {
			return api.Simulation{}, simulation.ErrNotFound
		},
	})

	_, err := useCase.Get(t.Context(), uuid.New())
	if !errors.Is(err, simulation.ErrNotFound) {
		t.Fatalf("Get() error = %v, want not found", err)
	}
}
