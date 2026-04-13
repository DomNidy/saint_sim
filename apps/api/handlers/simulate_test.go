//nolint:testpackage,exhaustruct
package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DomNidy/saint_sim/apps/api/middleware"
	dbqueries "github.com/DomNidy/saint_sim/pkg/db"
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
				APIKey: &dbqueries.GetApiKeyRow{
					PrincipalType: dbqueries.PrincipalTypeUser,
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
				APIKey: &dbqueries.GetApiKeyRow{
					PrincipalType: dbqueries.PrincipalTypeService,
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
			if ownerID.Valid != testCase.expectedOK {
				t.Fatalf("ownerID.Valid = %v, want %v", ownerID.Valid, testCase.expectedOK)
			}

			if ownerID.String != testCase.expectedID {
				t.Fatalf("ownerID.String = %q, want %q", ownerID.String, testCase.expectedID)
			}
		})
	}
}
