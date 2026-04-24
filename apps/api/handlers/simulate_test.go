//nolint:testpackage,exhaustruct
package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

type stubQueue struct {
	publish func(job simulation.JobMessage) error
}

func (q *stubQueue) Publish(job simulation.JobMessage) error {
	if q.publish != nil {
		return q.publish(job)
	}

	return nil
}

type stubSimulationStore struct {
	createQueuedSimulation func(
		context.Context,
		simulation.CreateQueuedSimulationInput,
	) (uuid.UUID, error)
	getSimulation func(context.Context, uuid.UUID) (api.Simulation, error)
}

const basicSimulationRequestBody = `{
	"kind":"basic",
	"core_config":{},
	"character":{
		"level":80,
		"character_class":"priest",
		"spec":"shadow",
		"race":"void_elf",
		"equipped_items":[]
	}
}`

const topGearSimulationRequestBody = `{
	"kind":"topGear",
	"core_config":{},
	"character":{
		"level":80,
		"character_class":"deathknight",
		"spec":"unholy",
		"role":"attack",
		"equipped_items":[],
		"active_talents":{"name":"Active","talents":"ACTIVE_TALENTS"}
	},
	"equipment":[
		{
			"slot":"head",
			"name":"Host Commander's Casque",
			"display_name":"Host Commander's Casque",
			"item_id":250458,
			"source":"equipped",
			"raw_line":"head=,id=250458,bonus_id=6652/12667/13577/13333/12787"
		},
		{
			"slot":"main_hand",
			"name":"Gnarlroot Spinecleaver",
			"display_name":"Gnarlroot Spinecleaver",
			"item_id":249671,
			"source":"equipped",
			"raw_line":"main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652"
		}
	]
}`

func (s stubSimulationStore) CreateQueuedSimulation(
	ctx context.Context,
	arg simulation.CreateQueuedSimulationInput,
) (uuid.UUID, error) {
	if s.createQueuedSimulation != nil {
		return s.createQueuedSimulation(ctx, arg)
	}

	return uuid.Nil, nil
}

func (s stubSimulationStore) GetSimulation(
	ctx context.Context,
	id uuid.UUID,
) (api.Simulation, error) {
	if s.getSimulation != nil {
		return s.getSimulation(ctx, id)
	}

	return api.Simulation{}, nil
}

func decodeTopGearSimulationRequestBody(t *testing.T) api.SimulationOptions {
	t.Helper()

	var requestBody api.SimulationOptions
	if err := json.Unmarshal([]byte(topGearSimulationRequestBody), &requestBody); err != nil {
		t.Fatalf("build top gear simulate request body: %v", err)
	}

	return requestBody
}

func decodeBasicSimulationRequestBody(t *testing.T) api.SimulationOptions {
	t.Helper()

	var requestBody api.SimulationOptions
	if err := json.Unmarshal([]byte(basicSimulationRequestBody), &requestBody); err != nil {
		t.Fatalf("build basic simulate request body: %v", err)
	}

	return requestBody
}

func assertBasicCreateSimulationParams(t *testing.T, arg simulation.CreateQueuedSimulationInput) {
	t.Helper()

	if arg.Kind != api.SimulationKindBasic {
		t.Fatalf("kind = %q, want %q", arg.Kind, api.SimulationKindBasic)
	}

	basicConfig, err := arg.Options.AsSimulationConfigBasic()
	if err != nil {
		t.Fatalf("read basic sim config: %v", err)
	}

	if basicConfig.Character.CharacterClass != api.Priest {
		t.Fatalf(
			"character_class = %q, want %q",
			basicConfig.Character.CharacterClass,
			api.Priest,
		)
	}
	if basicConfig.Character.Level != 80 {
		t.Fatalf("level = %d, want 80", basicConfig.Character.Level)
	}
	if basicConfig.Character.Spec != "shadow" {
		t.Fatalf("spec = %q, want shadow", basicConfig.Character.Spec)
	}
}

func assertTopGearCreateSimulationParams(t *testing.T, arg simulation.CreateQueuedSimulationInput) {
	t.Helper()

	if arg.Kind != api.SimulationKindTopGear {
		t.Fatalf("kind = %q, want %q", arg.Kind, api.SimulationKindTopGear)
	}

	topGearConfig, err := arg.Options.AsSimulationConfigTopGear()
	if err != nil {
		t.Fatalf("read top gear sim config: %v", err)
	}

	if topGearConfig.Character.CharacterClass != api.Deathknight {
		t.Fatalf(
			"character_class = %q, want %q",
			topGearConfig.Character.CharacterClass,
			api.Deathknight,
		)
	}
	if topGearConfig.Character.Spec != "unholy" {
		t.Fatalf("spec = %q, want %q", topGearConfig.Character.Spec, "unholy")
	}
	if topGearConfig.Character.Role == nil || *topGearConfig.Character.Role != "attack" {
		t.Fatalf("role = %v, want attack", topGearConfig.Character.Role)
	}
	if topGearConfig.Character.ActiveTalents.Talents != "ACTIVE_TALENTS" {
		t.Fatalf(
			"active_talents = %#v, want talents %q",
			topGearConfig.Character.ActiveTalents,
			"ACTIVE_TALENTS",
		)
	}
	if len(topGearConfig.Equipment) != 2 {
		t.Fatalf("equipment len = %d, want %d", len(topGearConfig.Equipment), 2)
	}
}

func assertAcceptedSimulationResponse(
	t *testing.T,
	response api.SimulateResponseObject,
	simulationID uuid.UUID,
) {
	t.Helper()

	acceptedResponse, ok := response.(api.Simulate202JSONResponse)
	if !ok {
		t.Fatalf("response type = %T, want %T", response, api.Simulate202JSONResponse{})
	}
	if acceptedResponse.SimulationId == nil ||
		*acceptedResponse.SimulationId != simulationID.String() {
		t.Fatalf(
			"simulation id = %v, want %s",
			acceptedResponse.SimulationId,
			simulationID.String(),
		)
	}
}

func testPublishSimulationJob(
	t *testing.T,
	requestBody api.SimulationOptions,
	assertCreateParams func(*testing.T, simulation.CreateQueuedSimulationInput),
) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	auth.SetAuthContext(ctx, auth.AuthContext{
		Scheme: auth.AuthSchemeBearer,
		UserID: "user-123",
	})

	didWriteToStore := false
	didPublishToQueue := false
	simulationID := uuid.New()

	server := NewServer(
		&stubSimulationStore{
			createQueuedSimulation: func(
				_ context.Context,
				arg simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				if didPublishToQueue {
					t.Fatal(
						"created simulation after published to queue. " +
							"this is incorrect order, want: create simulation, then publish to queue",
					)
				}

				assertCreateParams(t, arg)

				didWriteToStore = true

				return simulationID, nil
			},
		},
		&stubQueue{
			publish: func(_ simulation.JobMessage) error {
				if !didWriteToStore {
					t.Fatal(
						"got: published to queue before we created sim in store, want: create sim in store, then publish to queue",
					)
				}

				didPublishToQueue = true

				return nil
			},
		},
	)

	response, err := server.Simulate(ctx, api.SimulateRequestObject{
		Body: &requestBody,
	})
	if err != nil {
		t.Fatalf("Simulate() error = %v", err)
	}

	if !didPublishToQueue {
		t.Fatal("Simulate did not publish job to queue")
	}
	if !didWriteToStore {
		t.Fatal("Simulate did not write simulation to store")
	}

	assertAcceptedSimulationResponse(t, response, simulationID)
}

func TestPublishBasicSimulationJob(t *testing.T) {
	t.Parallel()

	testPublishSimulationJob(
		t,
		decodeBasicSimulationRequestBody(t),
		assertBasicCreateSimulationParams,
	)
}

func TestPublishTopGearSimulationJob(t *testing.T) {
	t.Parallel()

	testPublishSimulationJob(
		t,
		decodeTopGearSimulationRequestBody(t),
		assertTopGearCreateSimulationParams,
	)
}
