//nolint:testpackage,exhaustruct
package handlers

import (
	"net/http"
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestValidateSimulationRequest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		options    api.SimulationOptions
		wantStatus int
		wantOK     bool
	}{
		{
			name: "valid simc addon export",
			options: api.SimulationOptions{
				SimcAddonExport: "priest=\"Example\"\nlevel=80\nspec=shadow",
			},
			wantOK: true,
		},
		{
			name:       "missing simc addon export",
			options:    api.SimulationOptions{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty simc addon export",
			options: api.SimulationOptions{
				SimcAddonExport: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for idx := range cases {
		testCase := cases[idx]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validateSimulationRequest(t.Context(), testCase.options)
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
