//nolint:testpackage,exhaustruct
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseAddonExport(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	okRecorder := httptest.NewRecorder()
	okCtx, _ := gin.CreateTestContext(okRecorder)
	okCtx.Request = httptest.NewRequest(
		http.MethodPost,
		"/simc/parse-addon-export",
		bytes.NewBufferString(`{"simc_addon_export":"priest=\"Example\"\nlevel=80\nspec=shadow"}`),
	)
	okCtx.Request.Header.Set("Content-Type", "application/json")

	ParseAddonExport(okCtx)

	if okRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", okRecorder.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(okRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	addonExport, ok := payload["addon_export"].(map[string]any)
	if !ok {
		t.Fatalf("response missing addon_export: %v", payload)
	}

	if addonExport["class"] != "priest" {
		t.Fatalf("class = %v, want priest", addonExport["class"])
	}

	badRecorder := httptest.NewRecorder()
	badCtx, _ := gin.CreateTestContext(badRecorder)
	badCtx.Request = httptest.NewRequest(
		http.MethodPost,
		"/simc/parse-addon-export",
		bytes.NewBufferString(`{"simc_addon_export":"### comments only"}`),
	)
	badCtx.Request.Header.Set("Content-Type", "application/json")

	ParseAddonExport(badCtx)

	if badRecorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			badRecorder.Code,
			http.StatusBadRequest,
			badRecorder.Body.String(),
		)
	}
}
