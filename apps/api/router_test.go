//nolint:exhaustruct
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	"github.com/DomNidy/saint_sim/apps/api/handlers"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/simulation"
)

var (
	errUnexpectedToken   = errors.New("unexpected token")
	errUnexpectedAPIKey  = errors.New("unexpected api key")
	errInvalidTestToken  = errors.New("invalid bearer token")
	errInvalidTestAPIKey = errors.New("invalid api key")
)

type routerStubStore struct {
	createQueuedSimulation func(
		context.Context,
		simulation.CreateQueuedSimulationInput,
	) (uuid.UUID, error)
	getSimulation func(context.Context, uuid.UUID) (api.Simulation, error)
}

func (store routerStubStore) CreateQueuedSimulation(
	ctx context.Context,
	arg simulation.CreateQueuedSimulationInput,
) (uuid.UUID, error) {
	if store.createQueuedSimulation != nil {
		return store.createQueuedSimulation(ctx, arg)
	}

	return uuid.Nil, nil
}

func (store routerStubStore) GetSimulation(
	ctx context.Context,
	id uuid.UUID,
) (api.Simulation, error) {
	if store.getSimulation != nil {
		return store.getSimulation(ctx, id)
	}

	return api.Simulation{}, nil
}

type routerStubQueue struct {
	publish func(simulation.JobMessage) error
}

func (queue routerStubQueue) Publish(job simulation.JobMessage) error {
	if queue.publish != nil {
		return queue.publish(job)
	}

	return nil
}

type routerStubAuthenticator struct {
	authenticate func(context.Context, string) (auth.AuthContext, error)
}

func (auth routerStubAuthenticator) Authenticate(
	ctx context.Context,
	key string,
) (auth.AuthContext, error) {
	return auth.authenticate(ctx, key)
}

func TestRouterHealthEndpointRemainsReachable(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouterSimulationAuthAndValidation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	simulationID := uuid.MustParse("1b63b70e-e93d-4500-9bd7-c2644adf32f9")
	router := newTestRouter(
		t,
		withStore(routerStubStore{
			createQueuedSimulation: func(
				_ context.Context,
				_ simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				return simulationID, nil
			},
		}),
		withOpenAPIAuthenticator(auth.NewOpenAPIRequestAuthenticator(
			routerStubAuthenticator{
				authenticate: func(_ context.Context, rawToken string) (auth.AuthContext, error) {
					if rawToken != "valid-token" {
						return auth.AuthContext{}, errUnexpectedToken
					}

					return auth.AuthContext{
						Scheme: auth.AuthSchemeBearer,
						UserID: "user-123",
					}, nil
				},
			},
			routerStubAuthenticator{
				authenticate: func(_ context.Context, rawAPIKey string) (auth.AuthContext, error) {
					if rawAPIKey != "good-api-key" {
						return auth.AuthContext{}, errUnexpectedAPIKey
					}

					userID := "user-123"

					return auth.AuthContext{
						Scheme: auth.AuthSchemeAPIKey,
						APIKey: &db.GetApiKeyRow{
							PrincipalType: db.PrincipalTypeUser,
							UserID:        &userID,
						},
					}, nil
				},
			},
		)),
	)

	simConfig := api.SimulationConfigBasic{
		Kind: api.SimulationConfigBasicKindBasic,
		Character: api.WowCharacter{
			CharacterClass: api.Priest,
			EquippedItems:  []api.EquipmentItem{},
			Level:          80,
			Race:           "void_elf",
			Spec:           "shadow",
		},
		CoreConfig: api.SimulationCoreConfig{},
	}

	simOptionsJsonBody, err := json.Marshal(simConfig)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name       string
		headers    map[string]string
		body       string
		wantStatus int
	}{
		{
			name: "bearer auth succeeds",
			headers: map[string]string{
				"Authorization": "Bearer valid-token",
			},
			body:       string(simOptionsJsonBody),
			wantStatus: http.StatusAccepted,
		},
		{
			name: "api key fallback succeeds",
			headers: map[string]string{
				"Authorization": "Bearer bad-token",
				"Api-Key":       "good-api-key",
			},
			body:       string(simOptionsJsonBody),
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "missing credentials",
			body:       string(simOptionsJsonBody),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed bearer credentials",
			headers: map[string]string{
				"Authorization": "Bearer",
			},
			body:       string(simOptionsJsonBody),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing required field",
			headers: map[string]string{
				"Api-Key": "good-api-key",
			},
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for idx := range testCases {
		testCase := testCases[idx]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				http.MethodPost,
				"/simulation",
				bytes.NewBufferString(testCase.body),
			)
			req.Header.Set("Content-Type", "application/json")
			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != testCase.wantStatus {
				t.Fatalf(
					"status = %d, want %d; body = %s",
					rec.Code,
					testCase.wantStatus,
					rec.Body.String(),
				)
			}

			if testCase.wantStatus != http.StatusAccepted {
				return
			}

			var response map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if response["simulation_id"] != simulationID.String() {
				t.Fatalf(
					"simulation_id = %q, want %q",
					response["simulation_id"],
					simulationID.String(),
				)
			}
		})
	}
}

func TestRouterGeneratedBindingAndValidation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := newTestRouter(t)

	invalidIDReq := httptest.NewRequest(http.MethodGet, "/simulation/not-a-uuid", nil)
	invalidIDRec := httptest.NewRecorder()
	router.ServeHTTP(invalidIDRec, invalidIDReq)
	if invalidIDRec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			invalidIDRec.Code,
			http.StatusBadRequest,
			invalidIDRec.Body.String(),
		)
	}

}

type testRouterOptions struct {
	store                routerStubStore
	openAPIAuthenticator auth.OpenAPIRequestAuthenticator
}

type testRouterOption func(*testRouterOptions)

func withStore(store routerStubStore) testRouterOption {
	return func(options *testRouterOptions) {
		options.store = store
	}
}

func withOpenAPIAuthenticator(
	authenticator auth.OpenAPIRequestAuthenticator,
) testRouterOption {
	return func(options *testRouterOptions) {
		options.openAPIAuthenticator = authenticator
	}
}

func newTestRouter(t *testing.T, opts ...testRouterOption) *gin.Engine {
	t.Helper()

	options := testRouterOptions{
		store: routerStubStore{
			createQueuedSimulation: func(
				_ context.Context,
				_ simulation.CreateQueuedSimulationInput,
			) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		},
		openAPIAuthenticator: newTestOpenAPIAuthenticator(),
	}

	for _, opt := range opts {
		opt(&options)
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		t.Fatalf("GetSwagger() error = %v", err)
	}

	return newRouter(
		handlers.NewServer(options.store, routerStubQueue{}),
		swagger,
		options.openAPIAuthenticator,
	)
}

func newTestOpenAPIAuthenticator() auth.OpenAPIRequestAuthenticator {
	return auth.NewOpenAPIRequestAuthenticator(
		routerStubAuthenticator{
			authenticate: func(_ context.Context, rawToken string) (auth.AuthContext, error) {
				if rawToken != "valid-token" {
					return auth.AuthContext{}, errInvalidTestToken
				}

				return auth.AuthContext{
					Scheme: auth.AuthSchemeBearer,
					UserID: "user-123",
				}, nil
			},
		},
		routerStubAuthenticator{
			authenticate: func(_ context.Context, rawAPIKey string) (auth.AuthContext, error) {
				if rawAPIKey != "good-api-key" {
					return auth.AuthContext{}, errInvalidTestAPIKey
				}

				return auth.AuthContext{
					Scheme: auth.AuthSchemeAPIKey,
				}, nil
			},
		},
	)
}
