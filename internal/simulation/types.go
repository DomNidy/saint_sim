package simulation

import (
	"errors"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/internal/api"
)

// ErrNotFound indicates a simulation row was not found.
var ErrNotFound = errors.New("simulation not found")

// JSONMarshaler is implemented by structured artifacts that know how to
// encode themselves for database storage.
type JSONMarshaler interface {
	Marshal() ([]byte, error)
}

// CreateQueuedSimulationInput describes a simulation request to persist.
type CreateQueuedSimulationInput struct {
	Kind    api.SimulationKind
	Options api.SimulationOptions
	OwnerID *string
}

// SimulationRequest is the stored request shape loaded by the worker.
type SimulationRequest struct {
	ID      uuid.UUID
	Options api.SimulationOptions
}

// CompletedSimulation contains the artifacts written after a successful run.
type CompletedSimulation struct {
	Result             interface{}
	RawJSON2           JSONMarshaler
	RawSimcProfileText string
}

// FailedSimulation contains the user-facing failure text written for a failed run.
type FailedSimulation struct {
	ErrorText          string
	RawSimcProfileText string
}

// JobMessage is sent to the simulation queue. Worker consumes this to perform sims.
type JobMessage struct {
	SimulationID string `json:"simulation_id"`
}
