package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/internal/api_types"
	"github.com/DomNidy/saint_sim/internal/simc"
	"github.com/DomNidy/saint_sim/internal/utils"
)

const wowheadPreviewCacheTTL = 24 * time.Hour

//nolint:gochecknoglobals
var (
	buildGearPreview = simc.BuildGearPreview
	newWowheadClient = simc.NewWowheadXMLProvider
)

// SimcGearPreview parses a SimC addon export and returns grouped gear preview data.
func SimcGearPreview(ginContext *gin.Context) {
	var request api_types.GearPreviewRequest
	if err := ginContext.ShouldBindJSON(&request); err != nil {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("Invalid gear preview request"),
		})

		return
	}

	request.SimcAddonExport = simc.NormalizeLineEndings(request.SimcAddonExport)
	if strings.TrimSpace(request.SimcAddonExport) == "" {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("simc_addon_export is required"),
		})

		return
	}

	provider := newWowheadClient(nil, wowheadPreviewCacheTTL)
	groups, err := buildGearPreview(
		ginContext.Request.Context(),
		request.SimcAddonExport,
		provider,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		isBadRequest := strings.Contains(err.Error(), "no equipment") ||
			strings.Contains(err.Error(), "required")
		if isBadRequest {
			statusCode = http.StatusBadRequest
		}

		ginContext.JSON(statusCode, api_types.ErrorResponse{
			Message: utils.StrPtr(err.Error()),
		})

		return
	}

	ginContext.JSON(http.StatusOK, toGearPreviewResponse(groups))
}

func setGearPreviewDeps(
	builder func(context.Context, string, simc.ItemMetadataProvider) ([]simc.GearPreviewGroup, error),
	providerFactory func(*http.Client, time.Duration) simc.ItemMetadataProvider,
) func() {
	previousBuilder := buildGearPreview
	previousProviderFactory := newWowheadClient
	buildGearPreview = builder
	newWowheadClient = providerFactory

	return func() {
		buildGearPreview = previousBuilder
		newWowheadClient = previousProviderFactory
	}
}

func toGearPreviewResponse(groups []simc.GearPreviewGroup) api_types.GearPreviewResponse {
	responseGroups := make([]api_types.GearPreviewGroup, 0, len(groups))
	for idx := range groups {
		group := groups[idx]
		responseItems := make([]api_types.GearPreviewItem, 0, len(group.Items))

		for itemIdx := range group.Items {
			item := group.Items[itemIdx]
			itemLevel := item.ItemLevel
			responseItems = append(responseItems, api_types.GearPreviewItem{
				Fingerprint: item.Fingerprint,
				Slot:        item.Slot,
				Name:        item.Name,
				DisplayName: item.DisplayName,
				ItemId:      item.ItemID,
				ItemLevel:   itemLevel,
				IconUrl:     item.IconURL,
				WowheadUrl:  item.WowheadURL,
				WowheadData: item.WowheadData,
				Source:      api_types.GearPreviewItemSource(item.Source),
				RawLine:     item.RawLine,
			})
		}

		responseGroups = append(responseGroups, api_types.GearPreviewGroup{
			Slot:  group.Slot,
			Label: group.Label,
			Items: responseItems,
		})
	}

	return api_types.GearPreviewResponse{Groups: responseGroups}
}
