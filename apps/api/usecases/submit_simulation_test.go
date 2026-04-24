package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

var (
	errRepoBoom  = errors.New("repo boom")
	errQueueBoom = errors.New("queue boom")
)

type stubSubmitRepository struct {
	create func(context.Context, simulation.CreateQueuedSimulationInput) (uuid.UUID, error)
}

func (repo stubSubmitRepository) CreateQueuedSimulation(
	ctx context.Context,
	input simulation.CreateQueuedSimulationInput,
) (uuid.UUID, error) {
	return repo.create(ctx, input)
}

type stubSimulationQueue struct {
	publish func(simulation.JobMessage) error
}

func (queue stubSimulationQueue) Publish(message simulation.JobMessage) error {
	return queue.publish(message)
}

func TestSubmitSimulationCreatesBeforePublishing(t *testing.T) {
	t.Parallel()

	simulationID := uuid.New()
	var created bool
	useCase := NewSubmitSimulationUseCase(
		stubSubmitRepository{
			create: func(
				_ context.Context,
				input simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				if input.Kind != api.SimulationKindBasic {
					t.Fatalf("kind = %q, want %q", input.Kind, api.SimulationKindBasic)
				}

				created = true

				return simulationID, nil
			},
		},
		stubSimulationQueue{
			publish: func(message simulation.JobMessage) error {
				if !created {
					t.Fatal("published simulation job before creating queued simulation")
				}
				if message.SimulationID != simulationID.String() {
					t.Fatalf("simulation id = %q, want %q", message.SimulationID, simulationID)
				}

				return nil
			},
		},
	)

	gotID, err := useCase.Submit(t.Context(), SubmitSimulationInput{
		Options: basicSimulationOptions(t),
		OwnerID: stringPtr("user-123"),
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if gotID != simulationID {
		t.Fatalf("Submit() id = %s, want %s", gotID, simulationID)
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestSubmitSimulationPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	useCase := NewSubmitSimulationUseCase(
		stubSubmitRepository{
			create: func(
				context.Context,
				simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				return uuid.UUID{}, errRepoBoom
			},
		},
		stubSimulationQueue{
			publish: func(simulation.JobMessage) error {
				t.Fatal("Publish() called after repository error")
				return nil
			},
		},
	)

	_, err := useCase.Submit(t.Context(), SubmitSimulationInput{
		Options: basicSimulationOptions(t),
	})
	if !errors.Is(err, errRepoBoom) {
		t.Fatalf("Submit() error = %v, want repo error", err)
	}
}

func TestSubmitSimulationPropagatesPublishError(t *testing.T) {
	t.Parallel()

	useCase := NewSubmitSimulationUseCase(
		stubSubmitRepository{
			create: func(
				context.Context,
				simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		},
		stubSimulationQueue{
			publish: func(simulation.JobMessage) error {
				return errQueueBoom
			},
		},
	)

	_, err := useCase.Submit(t.Context(), SubmitSimulationInput{
		Options: basicSimulationOptions(t),
	})
	if !errors.Is(err, errQueueBoom) {
		t.Fatalf("Submit() error = %v, want queue error", err)
	}
}

func basicSimulationOptions(t *testing.T) api.SimulationOptions {
	t.Helper()

	var options api.SimulationOptions
	err := options.FromSimulationConfigBasic(api.SimulationConfigBasic{
		Kind: api.SimulationConfigBasicKindBasic,
		Character: api.WowCharacter{
			CharacterClass: api.Priest,
			EquippedItems:  []api.EquipmentItem{},
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
