package api

import (
	"context"
	"encoding/json"
	"testing"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
)

type fakeSimulationStore struct {
	createArg dbqueries.CreateSimulationRequestParams
	called    bool
}

func (f *fakeSimulationStore) CreateSimulationRequest(ctx context.Context, arg dbqueries.CreateSimulationRequestParams) error {
	f.called = true
	f.createArg = arg
	return nil
}

type fakeSimulationDispatcher struct {
	msg    api_types.SimulationMessageBody
	called bool
}

func (f *fakeSimulationDispatcher) DispatchSimulation(ctx context.Context, msg api_types.SimulationMessageBody) error {
	f.called = true
	f.msg = msg
	return nil
}

type fakeCharacterLookup struct {
	exists bool
}

func (f fakeCharacterLookup) Exists(character *api_types.WowCharacter) (bool, error) {
	return f.exists, nil
}

func TestSimulationServiceSubmitDispatchesSimulation(t *testing.T) {
	store := &fakeSimulationStore{}
	dispatcher := &fakeSimulationDispatcher{}
	service := NewSimulationService(
		store,
		dispatcher,
		fakeCharacterLookup{exists: true},
		func() string {
			return "11111111-1111-1111-1111-111111111111"
		},
	)

	response, err := service.Submit(context.Background(), api_types.SimulationOptions{
		WowCharacter: api_types.WowCharacter{
			CharacterName: "Testchar",
			Realm:         api_types.Draenor,
			Region:        api_types.Us,
		},
	})
	if err != nil {
		t.Fatalf("submit simulation: %v", err)
	}

	if response == nil || response.SimulationRequestId == nil {
		t.Fatal("response simulation id was nil")
	}

	if *response.SimulationRequestId != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("response simulation id = %q", *response.SimulationRequestId)
	}

	if !store.called {
		t.Fatal("expected store.CreateSimulationRequest to be called")
	}

	if !dispatcher.called {
		t.Fatal("expected dispatcher.DispatchSimulation to be called")
	}

	if dispatcher.msg.SimulationId == nil || *dispatcher.msg.SimulationId != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("dispatched simulation id = %#v", dispatcher.msg.SimulationId)
	}

	var storedOptions api_types.SimulationOptions
	if err := json.Unmarshal(store.createArg.Options, &storedOptions); err != nil {
		t.Fatalf("unmarshal stored options: %v", err)
	}

	if storedOptions.WowCharacter.CharacterName != "Testchar" {
		t.Fatalf("stored character name = %q", storedOptions.WowCharacter.CharacterName)
	}
}

func TestSimulationServiceSubmitReturnsNotFoundWhenCharacterMissing(t *testing.T) {
	service := NewSimulationService(
		&fakeSimulationStore{},
		&fakeSimulationDispatcher{},
		fakeCharacterLookup{exists: false},
		func() string {
			return "11111111-1111-1111-1111-111111111111"
		},
	)

	_, err := service.Submit(context.Background(), api_types.SimulationOptions{
		WowCharacter: api_types.WowCharacter{
			CharacterName: "Testchar",
			Realm:         api_types.Draenor,
			Region:        api_types.Us,
		},
	})
	if err != ErrCharacterNotFound {
		t.Fatalf("submit error = %v, want %v", err, ErrCharacterNotFound)
	}
}
