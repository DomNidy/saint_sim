package main

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
)

func simulationOptionsWithSimcConfig(t *testing.T, simcConfig string) api_types.SimulationOptions {
	t.Helper()

	var simOptions api_types.SimulationOptions
	value := reflect.ValueOf(&simOptions).Elem()
	simcConfigField := value.FieldByName("SimcConfig")
	if !simcConfigField.IsValid() {
		t.Skip("SimulationOptions.SimcConfig is not available yet")
	}

	if !simcConfigField.CanSet() || simcConfigField.Kind() != reflect.String {
		t.Fatalf("SimulationOptions.SimcConfig must be a settable string field")
	}

	simcConfigField.SetString(simcConfig)
	return simOptions
}

func TestSimulationServiceSubmitPersistsSanitizedSimcConfig(t *testing.T) {
	store := &fakeSimulationStore{}
	dispatcher := &fakeSimulationDispatcher{}
	service := SimulationService{
		store:      store,
		dispatcher: dispatcher,
		characters: fakeCharacterLookup{exists: true},
		idgen: func() string {
			return "11111111-1111-1111-1111-111111111111"
		},
	}

	rawSimcConfig := "# keep this comment out\niterations=9999\nmax_time=999\nvary_combat_length=0.7\nproxy=corp-proxy:8080\ndeathknight=\"John\"\ntalents+=alpha"
	wantSanitized := "iterations=500 max_time=300 vary_combat_length=0.1 deathknight=\"John\" talents+=alpha"

	response, err := service.Submit(context.Background(), simulationOptionsWithSimcConfig(t, rawSimcConfig))
	if err != nil {
		t.Fatalf("submit simulation: %v", err)
	}

	if response == nil || response.SimulationRequestId == nil {
		t.Fatal("response simulation id was nil")
	}

	if !store.called {
		t.Fatal("expected store.CreateSimulationRequest to be called")
	}

	wantJSON, err := json.Marshal(struct {
		SimcConfig string `json:"simc_config"`
	}{
		SimcConfig: wantSanitized,
	})
	if err != nil {
		t.Fatalf("marshal wanted options json: %v", err)
	}

	if !bytes.Equal(store.createArg.Options, wantJSON) {
		t.Fatalf("stored options json = %s, want %s", string(store.createArg.Options), string(wantJSON))
	}

	if bytes.Contains(store.createArg.Options, []byte("proxy=")) {
		t.Fatalf("stored options should not contain proxy: %s", string(store.createArg.Options))
	}

	if bytes.Contains(store.createArg.Options, []byte("iterations=9999")) {
		t.Fatalf("stored options should not contain unsanitized iterations: %s", string(store.createArg.Options))
	}

	if bytes.Contains(store.createArg.Options, []byte("max_time=999")) {
		t.Fatalf("stored options should not contain unsanitized max_time: %s", string(store.createArg.Options))
	}

	if bytes.Contains(store.createArg.Options, []byte("vary_combat_length=0.7")) {
		t.Fatalf("stored options should not contain unsanitized vary_combat_length: %s", string(store.createArg.Options))
	}
}
