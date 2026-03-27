package main

import (
	"context"
	"testing"

	pgtype "github.com/jackc/pgx/v5/pgtype"
)

func TestSimulationServiceSubmitDispatchesPersistedSimulationID(t *testing.T) {
	store := &fakeSimulationStore{}
	dispatcher := &fakeSimulationDispatcher{}

	const wantSimulationID = "11111111-1111-1111-1111-111111111111"

	service := SimulationService{
		store:      store,
		dispatcher: dispatcher,
		characters: fakeCharacterLookup{exists: true},
		idgen: func() string {
			return wantSimulationID
		},
	}

	response, err := service.Submit(context.Background(), simulationOptionsWithSimcConfig(t, "iterations=9999\ndeathknight=\"John\""))
	if err != nil {
		t.Fatalf("submit simulation: %v", err)
	}

	if !store.called {
		t.Fatal("expected store.CreateSimulationRequest to be called")
	}

	if !dispatcher.called {
		t.Fatal("expected dispatcher.DispatchSimulation to be called")
	}

	if response == nil || response.SimulationRequestId == nil {
		t.Fatal("response simulation id was nil")
	}

	if dispatcher.msg.SimulationId == nil {
		t.Fatal("dispatcher simulation id was nil")
	}

	if *dispatcher.msg.SimulationId != wantSimulationID {
		t.Fatalf("dispatched simulation id = %q, want %q", *dispatcher.msg.SimulationId, wantSimulationID)
	}

	if *response.SimulationRequestId != wantSimulationID {
		t.Fatalf("response simulation id = %q, want %q", *response.SimulationRequestId, wantSimulationID)
	}

	if *dispatcher.msg.SimulationId != *response.SimulationRequestId {
		t.Fatalf("dispatched simulation id = %q, response simulation id = %q", *dispatcher.msg.SimulationId, *response.SimulationRequestId)
	}

	var wantUUID pgtype.UUID
	if err := wantUUID.Scan(wantSimulationID); err != nil {
		t.Fatalf("scan expected simulation id into uuid: %v", err)
	}

	if store.createArg.ID.Valid != wantUUID.Valid || store.createArg.ID.Bytes != wantUUID.Bytes {
		t.Fatalf("persisted simulation id = %#v, want %#v", store.createArg.ID, wantUUID)
	}
}
