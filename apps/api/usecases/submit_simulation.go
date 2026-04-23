package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
	"github.com/DomNidy/saint_sim/internal/utils"
)

var (
	// ErrInvalidSimulationInput indicates semantic validation failed.
	ErrInvalidSimulationInput = errors.New("invalid simulation input")
	// ErrMalformedSimulationInput indicates the request cannot be decoded as a supported union.
	ErrMalformedSimulationInput = errors.New("malformed simulation input")
	// ErrUnsupportedSimulationInput indicates the request decoded to a kind the use case cannot
	// handle.
	ErrUnsupportedSimulationInput = errors.New("unsupported simulation input")
)

// SubmitSimulationRepository is the persistence boundary needed by the submit use case.
type SubmitSimulationRepository interface {
	CreateQueuedSimulation(
		ctx context.Context,
		input simulation.CreateQueuedSimulationInput,
	) (uuid.UUID, error)
}

// SimulationQueue publishes simulation jobs after they are created.
type SimulationQueue interface {
	Publish(simJob simulation.JobMessage) error
}

// SubmitSimulationInput contains the generated simulation options and resolved owner.
type SubmitSimulationInput struct {
	Options api.SimulationOptions
	OwnerID *string
}

// SubmitSimulationUseCase creates a queued simulation and publishes the worker job.
type SubmitSimulationUseCase struct {
	repository SubmitSimulationRepository
	queue      SimulationQueue
}

// NewSubmitSimulationUseCase constructs the submit use case.
func NewSubmitSimulationUseCase(
	repository SubmitSimulationRepository,
	queue SimulationQueue,
) *SubmitSimulationUseCase {
	return &SubmitSimulationUseCase{
		repository: repository,
		queue:      queue,
	}
}

// Submit validates input, persists the queued simulation, and publishes the queue message.
func (useCase *SubmitSimulationUseCase) Submit(
	ctx context.Context,
	input SubmitSimulationInput,
) (uuid.UUID, error) {
	kind, err := validateSimulationOptions(input.Options)
	if err != nil {
		return uuid.UUID{}, err
	}

	simulationID, err := useCase.repository.CreateQueuedSimulation(
		ctx,
		simulation.CreateQueuedSimulationInput{
			Kind:    kind,
			Options: input.Options,
			OwnerID: input.OwnerID,
		},
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create queued simulation: %w", err)
	}

	simulationJobMessage := simulation.JobMessage{
		SimulationID: simulationID.String(),
	}

	// TODO: Replace this direct DB-create-then-publish flow with a
	// transactional outbox so creating the simulation and recording the
	// publication intent happen in the same database transaction.
	err = useCase.queue.Publish(simulationJobMessage)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("publish simulation job: %w", err)
	}

	log.Printf(" [x] Sent %v\n", simulationJobMessage)

	return simulationID, nil
}

func validateSimulationOptions(options api.SimulationOptions) (api.SimulationKind, error) {
	config, err := options.ValueByDiscriminator()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformedSimulationInput, err)
	}

	switch simConfig := config.(type) {
	case api.SimulationConfigBasic:
		if err := utils.ValidateSimulationConfigBasic(&simConfig); err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidSimulationInput, err)
		}

		return api.SimulationKindBasic, nil
	case api.SimulationConfigTopGear:
		return api.SimulationKindTopGear, nil
	default:
		return "", fmt.Errorf("%w: %T", ErrUnsupportedSimulationInput, config)
	}
}
