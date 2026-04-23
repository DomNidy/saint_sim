package usecases

import (
	"context"

	"github.com/google/uuid"

	api "github.com/DomNidy/saint_sim/internal/api"
)

// GetSimulationRepository is the read boundary needed by the get simulation use case.
type GetSimulationRepository interface {
	GetSimulation(ctx context.Context, id uuid.UUID) (api.Simulation, error)
}

// GetSimulationUseCase loads simulation state for API callers.
type GetSimulationUseCase struct {
	repository GetSimulationRepository
}

// NewGetSimulationUseCase constructs the get simulation use case.
func NewGetSimulationUseCase(repository GetSimulationRepository) *GetSimulationUseCase {
	return &GetSimulationUseCase{
		repository: repository,
	}
}

// Get loads a simulation by id.
func (useCase *GetSimulationUseCase) Get(
	ctx context.Context,
	id uuid.UUID,
) (api.Simulation, error) {
	return useCase.repository.GetSimulation(ctx, id)
}
