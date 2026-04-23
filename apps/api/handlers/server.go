package handlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/usecases"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

type simulationQueue interface {
	Publish(simJob simulation.JobMessage) error
}

type submitSimulationUseCase interface {
	Submit(ctx context.Context, input usecases.SubmitSimulationInput) (uuid.UUID, error)
}

type getSimulationUseCase interface {
	Get(ctx context.Context, id uuid.UUID) (api.Simulation, error)
}

type simulationRepository interface {
	usecases.SubmitSimulationRepository
	usecases.GetSimulationRepository
}

// Server implements the generated strict OpenAPI server.
type Server struct {
	submitSimulation submitSimulationUseCase
	getSimulation    getSimulationUseCase
}

// NewServer constructs the strict OpenAPI server implementation.
func NewServer(repository simulationRepository, simQueue simulationQueue) *Server {
	return &Server{
		submitSimulation: usecases.NewSubmitSimulationUseCase(repository, simQueue),
		getSimulation:    usecases.NewGetSimulationUseCase(repository),
	}
}
