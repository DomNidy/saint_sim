//nolint:exhaustruct
package handlers

import (
	"encoding/json"
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestParseAddonExport(t *testing.T) {
	t.Parallel()

	server := NewServer(stubSimulationStore{}, &stubQueue{})

	okResponse, err := server.ParseAddonExport(
		t.Context(),
		api.ParseAddonExportRequestObject{
			Body: &api.ParseAddonExportRequest{
				SimcAddonExport: "priest=\"Example\"\nlevel=80\nspec=shadow",
			},
		},
	)
	if err != nil {
		t.Fatalf("ParseAddonExport() error = %v", err)
	}

	var payload map[string]any
	okBody, marshalErr := json.Marshal(okResponse)
	if marshalErr != nil {
		t.Fatalf("marshal response: %v", marshalErr)
	}
	if err := json.Unmarshal(okBody, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	addonExport, ok := payload["addon_export"].(map[string]any)
	if !ok {
		t.Fatalf("response missing addon_export: %v", payload)
	}

	if addonExport["class"] != "priest" {
		t.Fatalf("class = %v, want priest", addonExport["class"])
	}

	badResponse, err := server.ParseAddonExport(
		t.Context(),
		api.ParseAddonExportRequestObject{
			Body: &api.ParseAddonExportRequest{
				SimcAddonExport: "### comments only",
			},
		},
	)
	if err != nil {
		t.Fatalf("ParseAddonExport() error = %v", err)
	}

	if _, ok := badResponse.(api.ParseAddonExport400JSONResponse); !ok {
		t.Fatalf(
			"response type = %T, want %T",
			badResponse,
			api.ParseAddonExport400JSONResponse{},
		)
	}
}
