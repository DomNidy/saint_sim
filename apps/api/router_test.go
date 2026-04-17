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
	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/utils"
)

var (
	errUnexpectedToken   = errors.New("unexpected token")
	errUnexpectedAPIKey  = errors.New("unexpected api key")
	errInvalidTestToken  = errors.New("invalid bearer token")
	errInvalidTestAPIKey = errors.New("invalid api key")
)

type routerStubStore struct {
	createSimulation func(context.Context, db.CreateSimulationParams) (db.Simulation, error)
	getSimulation    func(context.Context, uuid.UUID) (db.Simulation, error)
}

func (store routerStubStore) CreateSimulation(
	ctx context.Context,
	arg db.CreateSimulationParams,
) (db.Simulation, error) {
	if store.createSimulation != nil {
		return store.createSimulation(ctx, arg)
	}

	return db.Simulation{}, nil
}

func (store routerStubStore) GetSimulation(
	ctx context.Context,
	id uuid.UUID,
) (db.Simulation, error) {
	if store.getSimulation != nil {
		return store.getSimulation(ctx, id)
	}

	return db.Simulation{}, nil
}

type routerStubQueue struct {
	publish func(utils.SimulationJobMessage) error
}

func (queue routerStubQueue) Publish(job utils.SimulationJobMessage) error {
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
			createSimulation: func(
				_ context.Context,
				_ db.CreateSimulationParams,
			) (db.Simulation, error) {
				return db.Simulation{ID: simulationID}, nil
			},
		}),
		withJWTAuthenticator(routerStubAuthenticator{
			authenticate: func(_ context.Context, rawToken string) (auth.AuthContext, error) {
				if rawToken != "valid-token" {
					return auth.AuthContext{}, errUnexpectedToken
				}

				return auth.AuthContext{
					Scheme: auth.AuthSchemeBearer,
					UserID: "user-123",
				}, nil
			},
		}),
		withAPIKeyAuthenticator(routerStubAuthenticator{
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
		}),
	)

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
			body:       `{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name: "api key fallback succeeds",
			headers: map[string]string{
				"Authorization": "Bearer bad-token",
				"Api-Key":       "good-api-key",
			},
			body:       `{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "missing credentials",
			body:       `{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed bearer credentials",
			headers: map[string]string{
				"Authorization": "Bearer",
			},
			body:       `{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`,
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

	parseReq := httptest.NewRequest(
		http.MethodPost,
		"/simc/parse-addon-export",
		bytes.NewBufferString(`{}`),
	)
	parseReq.Header.Set("Content-Type", "application/json")
	parseRec := httptest.NewRecorder()
	router.ServeHTTP(parseRec, parseReq)
	if parseRec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			parseRec.Code,
			http.StatusBadRequest,
			parseRec.Body.String(),
		)
	}
}

type testRouterOptions struct {
	store               routerStubStore
	jwtAuthenticator    auth.RequestAuthenticator
	apiKeyAuthenticator auth.RequestAuthenticator
}

type testRouterOption func(*testRouterOptions)

func withStore(store routerStubStore) testRouterOption {
	return func(options *testRouterOptions) {
		options.store = store
	}
}

func withJWTAuthenticator(auth auth.RequestAuthenticator) testRouterOption {
	return func(options *testRouterOptions) {
		options.jwtAuthenticator = auth
	}
}

func withAPIKeyAuthenticator(auth auth.RequestAuthenticator) testRouterOption {
	return func(options *testRouterOptions) {
		options.apiKeyAuthenticator = auth
	}
}

func newTestRouter(t *testing.T, opts ...testRouterOption) *gin.Engine {
	t.Helper()

	options := testRouterOptions{
		store: routerStubStore{
			createSimulation: func(
				_ context.Context,
				_ db.CreateSimulationParams,
			) (db.Simulation, error) {
				return db.Simulation{ID: uuid.New()}, nil
			},
		},
		jwtAuthenticator: routerStubAuthenticator{
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
		apiKeyAuthenticator: routerStubAuthenticator{
			authenticate: func(_ context.Context, rawAPIKey string) (auth.AuthContext, error) {
				if rawAPIKey != "good-api-key" {
					return auth.AuthContext{}, errInvalidTestAPIKey
				}

				return auth.AuthContext{
					Scheme: auth.AuthSchemeAPIKey,
				}, nil
			},
		},
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
		options.jwtAuthenticator,
		options.apiKeyAuthenticator,
	)
}
