//nolint:testpackage,exhaustruct
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

type stubQueue struct {
	publish func(job utils.SimulationJobMessage) error
}

func (q *stubQueue) Publish(job utils.SimulationJobMessage) error {
	if q.publish != nil {
		return q.publish(job)
	}

	return nil
}

type stubSimulationStore struct {
	createSimulation func(context.Context, db.CreateSimulationParams) (db.Simulation, error)
	getSimulation    func(context.Context, uuid.UUID) (db.Simulation, error)
}

const topGearSimulationRequestBody = `{
	"kind":"topGear",
	"character_name":"Gubulgi",
	"class":"deathknight",
	"spec":"unholy",
	"role":"attack",
	"talent_loadout":{"name":"Active","talents":"ACTIVE_TALENTS"},
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

func (s stubSimulationStore) CreateSimulation(
	ctx context.Context,
	arg db.CreateSimulationParams,
) (db.Simulation, error) {
	if s.createSimulation != nil {
		return s.createSimulation(ctx, arg)
	}

	return db.Simulation{}, nil
}

func (s stubSimulationStore) GetSimulation(
	ctx context.Context,
	id uuid.UUID,
) (db.Simulation, error) {
	if s.getSimulation != nil {
		return s.getSimulation(ctx, id)
	}

	return db.Simulation{}, nil
}

func decodeTopGearSimulationRequestBody(t *testing.T) api.SimulationOptions {
	t.Helper()

	var requestBody api.SimulationOptions
	if err := json.Unmarshal([]byte(topGearSimulationRequestBody), &requestBody); err != nil {
		t.Fatalf("build top gear simulate request body: %v", err)
	}

	return requestBody
}

func assertTopGearCreateSimulationParams(t *testing.T, arg db.CreateSimulationParams) {
	t.Helper()

	if arg.Kind != db.SimulationKindTopGear {
		t.Fatalf("kind = %q, want %q", arg.Kind, db.SimulationKindTopGear)
	}

	var topGearConfig api.SimulationConfigTopGear
	if err := json.Unmarshal(arg.SimConfig, &topGearConfig); err != nil {
		t.Fatalf("unmarshal top gear sim config: %v", err)
	}

	if topGearConfig.CharacterName != "Gubulgi" {
		t.Fatalf(
			"character_name = %q, want %q",
			topGearConfig.CharacterName,
			"Gubulgi",
		)
	}
	if topGearConfig.Class != api.Deathknight {
		t.Fatalf("class = %q, want %q", topGearConfig.Class, api.Deathknight)
	}
	if topGearConfig.Spec != "unholy" {
		t.Fatalf("spec = %q, want %q", topGearConfig.Spec, "unholy")
	}
	if topGearConfig.TalentLoadout.Talents != "ACTIVE_TALENTS" {
		t.Fatalf(
			"talents = %q, want %q",
			topGearConfig.TalentLoadout.Talents,
			"ACTIVE_TALENTS",
		)
	}
	if len(topGearConfig.Equipment) != 2 {
		t.Fatalf("equipment len = %d, want %d", len(topGearConfig.Equipment), 2)
	}
}

//nolint:cyclop
func TestPublishSimulationJob(t *testing.T) {
	t.Parallel()
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
	expectedExport := "priest=\"Example\"\nlevel=80\nspec=shadow"

	server := NewServer(
		&stubSimulationStore{
			createSimulation: func(
				_ context.Context,
				arg db.CreateSimulationParams,
			) (db.Simulation, error) {
				if didPublishToQueue {
					t.Fatal(
						"created simulation after published to queue. " +
							"this is incorrect order, want: create simulation, then publish to queue",
					)
				}

				var simOptions api.SimulationOptions
				if err := json.Unmarshal(arg.SimConfig, &simOptions); err != nil {
					t.Fatalf("unmarshal sim config: %v", err)
				}
				basicConfig, err := simOptions.AsSimulationConfigBasic()
				if err != nil {
					t.Fatalf("decode basic sim options: %v", err)
				}
				if basicConfig.SimcAddonExport != expectedExport {
					t.Fatalf(
						"simc_addon_export = %q, want %q",
						basicConfig.SimcAddonExport,
						expectedExport,
					)
				}

				didWriteToStore = true

				return db.Simulation{ID: simulationID}, nil
			},
		},
		&stubQueue{
			publish: func(_ utils.SimulationJobMessage) error {
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
	var requestBody api.SimulationOptions
	if err := json.Unmarshal([]byte(`{
		"kind":"basic",
		"simc_addon_export":"priest=\"Example\"\r\nlevel=80\rspec=shadow"
	}`), &requestBody); err != nil {
		t.Fatalf("build simulate request body: %v", err)
	}

	response, err := server.Simulate(ctx, api.SimulateRequestObject{
		Body: &requestBody,
	})
	if err != nil {
		t.Fatalf("Simulate() error = %v", err)
	}

	if !didPublishToQueue {
		t.Fatal("Simulate did not publish to queue")
	}
	if !didWriteToStore {
		t.Fatal("Simulate did not write to store")
	}

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

func TestPublishTopGearSimulationJob(t *testing.T) {
	t.Parallel()
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
			createSimulation: func(
				_ context.Context,
				arg db.CreateSimulationParams,
			) (db.Simulation, error) {
				if didPublishToQueue {
					t.Fatal(
						"created simulation after published to queue. " +
							"this is incorrect order, want: create simulation, then publish to queue",
					)
				}

				assertTopGearCreateSimulationParams(t, arg)

				didWriteToStore = true

				return db.Simulation{ID: simulationID}, nil
			},
		},
		&stubQueue{
			publish: func(_ utils.SimulationJobMessage) error {
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

	requestBody := decodeTopGearSimulationRequestBody(t)

	response, err := server.Simulate(ctx, api.SimulateRequestObject{
		Body: &requestBody,
	})
	if err != nil {
		t.Fatalf("Simulate() error = %v", err)
	}

	if !didPublishToQueue {
		t.Fatal("Simulate did not publish top gear job to queue")
	}
	if !didWriteToStore {
		t.Fatal("Simulate did not write top gear simulation to store")
	}

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

func TestSimulationOwnerID(t *testing.T) {
	t.Parallel()

	userID := "user-123"

	tests := []struct {
		name        string
		authContext auth.AuthContext
		want        *string
	}{
		{
			name: "bearer auth uses subject",
			authContext: auth.AuthContext{
				Scheme: auth.AuthSchemeBearer,
				UserID: userID,
			},
			want: &userID,
		},
		{
			name: "user-owned api key uses principal user id",
			authContext: auth.AuthContext{
				Scheme: auth.AuthSchemeAPIKey,
				APIKey: &db.GetApiKeyRow{
					PrincipalType: db.PrincipalTypeUser,
					UserID:        &userID,
				},
			},
			want: &userID,
		},
		{
			name: "service-owned api key remains unowned",
			authContext: auth.AuthContext{
				Scheme: auth.AuthSchemeAPIKey,
				APIKey: &db.GetApiKeyRow{
					PrincipalType: db.PrincipalTypeService,
				},
			},
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := simulationOwnerID(test.authContext)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("simulationOwnerID() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateSimulationRequest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		simConfig  api.SimulationConfigBasic
		wantStatus int
		wantOK     bool
	}{
		{
			name: "valid simc addon export",
			simConfig: api.SimulationConfigBasic{
				SimcAddonExport: "priest=\"Example\"\nlevel=80\nspec=shadow",
			},
			wantOK: true,
		},
		{
			name:       "missing simc addon export",
			simConfig:  api.SimulationConfigBasic{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty simc addon export",
			simConfig: api.SimulationConfigBasic{
				SimcAddonExport: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for idx := range cases {
		testCase := cases[idx]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validateSimulationRequestBasic(t.Context(), testCase.simConfig)
			if testCase.wantOK {
				if result != nil {
					t.Fatalf("validateSimulationRequest() = %#v, want nil", result)
				}

				return
			}

			if result == nil {
				t.Fatal("validateSimulationRequest() = nil, want error")
			}

			if result.statusCode != testCase.wantStatus {
				t.Fatalf("statusCode = %d, want %d", result.statusCode, testCase.wantStatus)
			}
		})
	}
}
