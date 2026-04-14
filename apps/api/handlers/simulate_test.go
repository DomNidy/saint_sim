//nolint:testpackage,exhaustruct
package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/middleware"
	"github.com/DomNidy/saint_sim/pkg/db"
	"github.com/DomNidy/saint_sim/pkg/utils"
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

func TestPublishSimulationJob(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/simulation",
		bytes.NewBufferString(`{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	didWriteToStore := false
	didPublishToQueue := false
	simulationID := uuid.New()

	Simulate(
		ctx,
		&stubSimulationStore{
			createSimulation: func(
				_ context.Context,
				_ db.CreateSimulationParams,
			) (db.Simulation, error) {
				if didPublishToQueue {
					t.Fatal(
						"created simulation after published to queue. " +
							"this is incorrect order, want: create simulation, then publish to queue",
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

	if !didPublishToQueue {
		t.Fatal("Simulate did not publish to queue")
	}
	if !didWriteToStore {
		t.Fatal("Simulate did not write to store")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			http.StatusAccepted,
			rec.Body.String(),
		)
	}
}
