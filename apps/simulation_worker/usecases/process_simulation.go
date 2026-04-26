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
	WriteRunDetails(ctx context.Context, id uuid.UUID, rawProfileText string) error
}

// Runner is the abstraction that orchestrates the execution of simc
// against a given sim Plan.
type Runner interface {
	Run(
		ctx context.Context,
		plan sims.Plan,
	) (api.SimulationResult, error)
}

// ProcessSimulationUseCase processes one queued simulation id.
type ProcessSimulationUseCase struct {
	repo   SimulationRepository
	runner Runner
}

// NewProcessSimulationUseCase constructs a process use case with an injected runner.
func NewProcessSimulationUseCase(
	repository SimulationRepository,
	runner Runner,
) ProcessSimulationUseCase {
	return ProcessSimulationUseCase{
		repo:   repository,
		runner: runner,
	}
}

// Process loads, runs, and marks one simulation according to current worker semantics.
func (useCase *ProcessSimulationUseCase) Process(ctx context.Context, requestID uuid.UUID) error {
	request, err := useCase.repo.LoadRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("load simulation request: %w", err)
	}

	log.Printf("loaded request from repo: %v", request.ID)

	// read the "kind" field to determine what kind of simulation job
	// it is.
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

	// route the simulation job to the correct handler for its kind
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

	if err := useCase.repo.MarkInProgress(ctx, request.ID); err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.ID.String(), err)
	}

	plan, err := sims.NewBasicSimPlan(config)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}

	rawProfileText, err := plan.BuildSimcProfile()
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}
	useCase.repo.WriteRunDetails(ctx, request.ID, string(rawProfileText))

	result, err := useCase.runner.Run(ctx, plan)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("run simulation: %w", err))
	}

	err = useCase.repo.MarkCompleted(ctx, request.ID, simulation.CompletedSimulation{
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

	plan, err := sims.NewTopGearSimPlan(opts)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}

	rawProfileText, err := plan.BuildSimcProfile()
	if err != nil {
		return useCase.failRequest(ctx, request.ID, err)
	}
	useCase.repo.WriteRunDetails(ctx, request.ID, string(rawProfileText))

	if err := useCase.repo.MarkInProgress(ctx, request.ID); err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.ID.String(), err)
	}

	result, err := useCase.runner.Run(ctx, plan)
	if err != nil {
		return useCase.failRequest(ctx, request.ID, fmt.Errorf("run top gear simulation: %w", err))
	}

	err = useCase.repo.MarkCompleted(ctx, request.ID, simulation.CompletedSimulation{
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
	return useCase.repo.MarkFailed(ctx, id, simulation.FailedSimulation{
		ErrorText: "internal server error",
	})
}
