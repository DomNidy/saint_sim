package handlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

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
