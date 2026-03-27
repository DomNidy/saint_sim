package main

import (
	"context"
	"encoding/json"
	"testing"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeStore struct {
	optionsByID map[string][]byte
	inserted    map[string]string
}

func (f *fakeStore) GetSimulationRequestOptions(ctx context.Context, id pgtype.UUID) ([]byte, error) {
	key, err := idValue(id)
	if err != nil {
		return nil, err
	}
	return f.optionsByID[key], nil
}

func (f *fakeStore) InsertSimulationData(ctx context.Context, arg dbqueries.InsertSimulationDataParams) error {
	if f.inserted == nil {
		f.inserted = map[string]string{}
	}
	key, err := idValue(arg.RequestID)
	if err != nil {
		return err
	}
	f.inserted[key] = arg.SimResult
	return nil
}

type fakeRunner struct {
	region string
	realm  string
	name   string
	result []byte
}

func (f *fakeRunner) Perform(region, realm, name string) ([]byte, error) {
	f.region = region
	f.realm = realm
	f.name = name
	return f.result, nil
}

func (f *fakeRunner) Version() string {
	return "test"
}

func TestWorkerHandleMessageStoresSimulationResult(t *testing.T) {
	const requestIDText = "11111111-1111-1111-1111-111111111111"
	requestID := pgtype.UUID{}
	if err := requestID.Scan(requestIDText); err != nil {
		t.Fatalf("scan request id: %v", err)
	}

	simOptionsJSON, err := json.Marshal(api_types.SimulationOptions{
		WowCharacter: api_types.WowCharacter{
			CharacterName: "Testchar",
			Realm:         api_types.Draenor,
			Region:        api_types.Us,
		},
	})
	if err != nil {
		t.Fatalf("marshal simulation options: %v", err)
	}

	store := &fakeStore{
		optionsByID: map[string][]byte{
			requestIDText: simOptionsJSON,
		},
	}
	runner := &fakeRunner{result: []byte("sim output")}
	worker := Worker{
		store:  store,
		runner: runner,
	}

	msgBody, err := json.Marshal(api_types.SimulationMessageBody{
		SimulationId: strPtr(requestIDText),
	})
	if err != nil {
		t.Fatalf("marshal message body: %v", err)
	}

	if err := worker.HandleMessage(context.Background(), msgBody); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	if runner.region != string(api_types.Us) {
		t.Fatalf("runner region = %q, want %q", runner.region, api_types.Us)
	}

	if runner.realm != string(api_types.Draenor) {
		t.Fatalf("runner realm = %q, want %q", runner.realm, api_types.Draenor)
	}

	if runner.name != "Testchar" {
		t.Fatalf("runner name = %q, want %q", runner.name, "Testchar")
	}

	got := store.inserted[requestIDText]
	if got != "sim output" {
		t.Fatalf("inserted sim result = %q, want %q", got, "sim output")
	}
}

func idValue(id pgtype.UUID) (string, error) {
	value, err := id.Value()
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func strPtr(s string) *string {
	return &s
}
