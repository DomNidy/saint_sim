package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/sims"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

var errRunBoom = errors.New("run boom")

type stubWorkerRepository struct {
	request    simulation.SimulationRequest
	loadErr    error
	markErr    error
	completed  bool
	failed     bool
	inProgress bool
}

func (repo *stubWorkerRepository) LoadRequest(
	context.Context,
	uuid.UUID,
) (simulation.SimulationRequest, error) {
	return repo.request, repo.loadErr
}

func (repo *stubWorkerRepository) MarkInProgress(context.Context, uuid.UUID) error {
	repo.inProgress = true

	return repo.markErr
}

func (repo *stubWorkerRepository) MarkCompleted(
	context.Context,
	uuid.UUID,
	simulation.CompletedSimulation,
) error {
	repo.completed = true

	return repo.markErr
}

func (repo *stubWorkerRepository) MarkFailed(
	context.Context,
	uuid.UUID,
	simulation.FailedSimulation,
) error {
	repo.failed = true

	return nil
}

func (repo *stubWorkerRepository) WriteRunDetails(
	ctx context.Context,
	id uuid.UUID,
	rawProfileText string,
) error {
	return nil
}

type stubRunner struct {
	plan sims.Plan
	err  error
}

func (runner *stubRunner) Run(
	_ context.Context,
	plan sims.Plan,
) (api.SimulationResult, error) {
	runner.plan = plan
	if runner.err != nil {
		return api.SimulationResult{}, runner.err
	}

	return api.SimulationResult{}, nil
}

func TestProcessSimulationMarksBasicSimulationCompleted(t *testing.T) {
	t.Parallel()

	requestID := uuid.New()
	repo := &stubWorkerRepository{
		request: simulation.SimulationRequest{
			ID:      requestID,
			Options: basicWorkerSimulationOptions(t),
		},
	}
	runner := &stubRunner{}
	useCase := NewProcessSimulationUseCase(repo, runner)

	if err := useCase.Process(t.Context(), requestID); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if !repo.inProgress {
		t.Fatal("simulation was not marked in progress")
	}
	if !repo.completed {
		t.Fatal("simulation was not marked complete")
	}
	if repo.failed {
		t.Fatal("simulation was marked failed")
	}
	if _, ok := runner.plan.(sims.BasicSimPlan); !ok {
		t.Fatalf("plan type = %T, want BasicSimPlan", runner.plan)
	}
}

func TestProcessSimulationMarksFailedWhenRunFails(t *testing.T) {
	t.Parallel()

	requestID := uuid.New()
	repo := &stubWorkerRepository{
		request: simulation.SimulationRequest{
			ID:      requestID,
			Options: basicWorkerSimulationOptions(t),
		},
	}
	useCase := NewProcessSimulationUseCase(repo, &stubRunner{err: errRunBoom})

	err := useCase.Process(t.Context(), requestID)
	if !errors.Is(err, errRunBoom) {
		t.Fatalf("Process() error = %v, want run error", err)
	}
	if !repo.failed {
		t.Fatal("simulation was not marked failed")
	}
}

func TestProcessSimulationPropagatesLoadError(t *testing.T) {
	t.Parallel()

	repo := &stubWorkerRepository{loadErr: simulation.ErrNotFound}
	useCase := NewProcessSimulationUseCase(repo, &stubRunner{})

	err := useCase.Process(t.Context(), uuid.New())
	if !errors.Is(err, simulation.ErrNotFound) {
		t.Fatalf("Process() error = %v, want not found", err)
	}
}

func basicWorkerSimulationOptions(t *testing.T) api.SimulationOptions {
	t.Helper()

	var options api.SimulationOptions
	err := options.FromSimulationConfigBasic(api.SimulationConfigBasic{
		Kind: api.SimulationConfigBasicKindBasic,
		Character: api.WowCharacter{
			CharacterClass: api.Priest,
			EquippedItems:  api.CharacterEquippedItems{},
			Level:          80,
			Race:           "void_elf",
			Spec:           "shadow",
		},
		CoreConfig: api.SimulationCoreConfig{},
	})
	if err != nil {
		t.Fatalf("encode basic simulation options: %v", err)
	}

	return options
}
