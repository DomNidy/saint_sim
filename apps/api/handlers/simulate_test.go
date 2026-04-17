//nolint:testpackage,exhaustruct
package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/middleware"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

func TestSimulationOwnerID(t *testing.T) {
	t.Parallel()

	userID := "user-42"
	serviceID := uuid.MustParse("20875e8d-a145-4310-b503-89f0884c5008")

	cases := []struct {
		name        string
		authContext middleware.AuthContext
		expectedID  string
		expectedOK  bool
	}{
		{
			name: "bearer auth",
			authContext: middleware.AuthContext{
				Scheme: middleware.AuthSchemeBearer,
				UserID: userID,
			},
			expectedID: userID,
			expectedOK: true,
		},
		{
			name: "user-owned api key",
			authContext: middleware.AuthContext{
				Scheme: middleware.AuthSchemeAPIKey,
				APIKey: &db.GetApiKeyRow{
					PrincipalType: db.PrincipalTypeUser,
					UserID:        &userID,
				},
			},
			expectedID: userID,
			expectedOK: true,
		},
		{
			name: "service-owned api key",
			authContext: middleware.AuthContext{
				Scheme: middleware.AuthSchemeAPIKey,
				APIKey: &db.GetApiKeyRow{
					PrincipalType: db.PrincipalTypeService,
					ServiceID:     &serviceID,
				},
			},
			expectedID: "",
			expectedOK: false,
		},
	}

	for idx := range cases {
		testCase := cases[idx]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ginContext := &gin.Context{}
			ginContext.Set("auth.context", testCase.authContext)

			ownerID := simulationOwnerID(ginContext)
			if (ownerID != nil) != testCase.expectedOK {
				t.Fatalf("ownerID != nil = %v, want %v", ownerID != nil, testCase.expectedOK)
			}

			if ownerID == nil {
				if testCase.expectedID != "" {
					t.Fatalf("ownerID = nil, want %q", testCase.expectedID)
				}

				return
			}

			if *ownerID != testCase.expectedID {
				t.Fatalf("*ownerID = %q, want %q", *ownerID, testCase.expectedID)
			}
		})
	}
}

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

//nolint:cyclop
func TestPublishSimulationJob(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ctx := &gin.Context{}

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
				if simOptions.SimcAddonExport != expectedExport {
					t.Fatalf(
						"simc_addon_export = %q, want %q",
						simOptions.SimcAddonExport,
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
	response, err := server.Simulate(ctx, api.SimulateRequestObject{
		Body: &api.SimulationOptions{
			SimcAddonExport: "priest=\"Example\"\r\nlevel=80\rspec=shadow",
		},
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
