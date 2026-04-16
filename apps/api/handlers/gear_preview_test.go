//nolint:testpackage,exhaustruct
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/internal/simc"
)

var errNoEquipment = errors.New("no equipment lines found in addon export")

type noopMetadataProvider struct{}

func (noopMetadataProvider) Metadata(context.Context, int) (simc.WowheadItemMetadata, error) {
	return simc.WowheadItemMetadata{}, nil
}

func TestSimcGearPreview(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	teardown := setGearPreviewDeps(
		func(_ context.Context, input string, _ simc.ItemMetadataProvider) ([]simc.GearPreviewGroup, error) {
			if input == "bad" {
				return nil, errNoEquipment
			}

			return []simc.GearPreviewGroup{
				{
					Slot:  "head",
					Label: "Head",
					Items: []simc.GearPreviewItem{{
						Fingerprint: "fp",
						Slot:        "head",
						Name:        "Foo",
						DisplayName: "Foo",
						ItemID:      123,
						WowheadURL:  "https://www.wowhead.com/item=123",
						WowheadData: "item=123",
						Source:      "equipped",
						RawLine:     "head=,id=123",
					}},
				},
			}, nil
		},
		func(*http.Client, time.Duration) simc.ItemMetadataProvider {
			return noopMetadataProvider{}
		},
	)
	t.Cleanup(teardown)

	okRecorder := httptest.NewRecorder()
	okCtx, _ := gin.CreateTestContext(okRecorder)
	okCtx.Request = httptest.NewRequest(
		http.MethodPost,
		"/simc/gear-preview",
		bytes.NewBufferString(`{"simc_addon_export":"head=,id=123"}`),
	)
	okCtx.Request.Header.Set("Content-Type", "application/json")

	SimcGearPreview(okCtx)

	if okRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", okRecorder.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(okRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if _, ok := payload["groups"]; !ok {
		t.Fatalf("response missing groups: %v", payload)
	}

	badRecorder := httptest.NewRecorder()
	badCtx, _ := gin.CreateTestContext(badRecorder)
	badCtx.Request = httptest.NewRequest(
		http.MethodPost,
		"/simc/gear-preview",
		bytes.NewBufferString(`{"simc_addon_export":"bad"}`),
	)
	badCtx.Request.Header.Set("Content-Type", "application/json")

	SimcGearPreview(badCtx)

	if badRecorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			badRecorder.Code,
			http.StatusBadRequest,
			badRecorder.Body.String(),
		)
	}
}
