package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/sims"
	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
	"github.com/DomNidy/saint_sim/internal/utils"
)

// ErrUnsupportedSimulationKind indicates the worker does not support the stored kind.
var ErrUnsupportedSimulationKind = errors.New("unsupported simulation kind")

// SimulationRepository is the worker-owned persistence interface.
//
// Future reliability work should replace the separate LoadRequest and
// MarkInProgress calls with an atomic claim method, for example a method that
// updates only rows currently in_queue and returns whether this worker won the
// right to process the simulation.
type SimulationRepository interface {
	LoadRequest(ctx context.Context, requestID uuid.UUID) (simulation.SimulationRequest, error)
	MarkInProgress(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID, result simulation.CompletedSimulation) error
	MarkFailed(ctx context.Context, id uuid.UUID, failure simulation.FailedSimulation) error
}

// Runner is the abstraction that orchestrates the execution of simc
// against a given sim Manifest.
type Runner interface {
	Run(
		ctx context.Context,
		manifest sims.Manifest,
	) (api.SimulationResult, error)
}

// ProcessSimulationUseCase processes one queued simulation id.
type ProcessSimulationUseCase struct {
	repository SimulationRepository
	runner     Runner
}

// NewProcessSimulationUseCase constructs a process use case with an injected runner.
func NewProcessSimulationUseCase(
	repository SimulationRepository,
	runner Runner,
) *ProcessSimulationUseCase {
	return &ProcessSimulationUseCase{
		repository: repository,
		runner:     runner,
	}
}

// Process loads, runs, and marks one simulation according to current worker semantics.
func (useCase *ProcessSimulationUseCase) Process(ctx context.Context, requestID uuid.UUID) error {
	request, err := useCase.repository.LoadRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("load simulation request: %w", err)
	}

	kind, err := request.Options.Discriminator()
	if err != nil {
		if markErr := useCase.markFailed(ctx, requestID); markErr != nil {
			return errors.Join(
				fmt.Errorf("read simulation discriminator: %w", err),
				markErr,
			)
		}

		return fmt.Errorf("read simulation discriminator: %w", err)
	}

	switch kind {
	case string(api.SimulationKindBasic):
		return useCase.processBasic(ctx, request)
	case string(api.SimulationKindTopGear):
		return useCase.processTopGear(ctx, request)
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedSimulationKind, kind)
	}
}

func (useCase *ProcessSimulationUseCase) processBasic(
	ctx context.Context,
	request simulation.SimulationRequest,
) error {
	config, err := request.Options.AsSimulationConfigBasic()
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("cast to basic: %w", err))
	}

	if err := utils.ValidateSimulationConfigBasic(&config); err != nil {
		return useCase.failRequest(
			ctx,
			request.ID,
			fmt.Errorf("validate simulation options: %w", err),
		)
	}

	if err := useCase.repository.MarkInProgress(ctx, request.ID); err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.ID.String(), err)
	}

	manifest, err := sims.NewBasicSimManifest(config)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}

	result, err := useCase.runner.Run(ctx, manifest)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("run simulation: %w", err))
	}

	err = useCase.repository.MarkCompleted(ctx, request.ID, simulation.CompletedSimulation{
		Result: result,
	})
	if err != nil {
		return fmt.Errorf("persist simulation result: %w", err)
	}

	return nil
}

func (useCase *ProcessSimulationUseCase) processTopGear(
	ctx context.Context,
	request simulation.SimulationRequest,
) error {
	opts, err := request.Options.AsSimulationConfigTopGear()
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("cast to topGear: %w", err))
	}

	manifest, err := sims.NewTopGearManifest(opts)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}

	if err := useCase.repository.MarkInProgress(ctx, request.ID); err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.ID.String(), err)
	}

	result, err := useCase.runner.Run(ctx, manifest)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("run top gear simulation: %w", err))
	}

	err = useCase.repository.MarkCompleted(ctx, request.ID, simulation.CompletedSimulation{
		Result: result,
	})
	if err != nil {
		return fmt.Errorf("persist simulation result: %w", err)
	}

	return nil
}

func (useCase *ProcessSimulationUseCase) failRequest(
	ctx context.Context,
	id uuid.UUID,
	cause error,
) error {
	if markErr := useCase.markFailed(ctx, id); markErr != nil {
		return errors.Join(cause, markErr)
	}

	return cause
}

func (useCase *ProcessSimulationUseCase) markFailed(ctx context.Context, id uuid.UUID) error {
	return useCase.repository.MarkFailed(ctx, id, simulation.FailedSimulation{
		ErrorText: "internal server error",
	})
}
