package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DomNidy/saint_sim/internal/api_types"
	"github.com/DomNidy/saint_sim/internal/simc"
	"github.com/DomNidy/saint_sim/internal/utils"
)

// ParseAddonExport parses a SimC addon export and returns structured data.
func ParseAddonExport(ginContext *gin.Context) {
	var request api_types.ParseAddonExportRequest
	if err := ginContext.ShouldBindJSON(&request); err != nil {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("Invalid parse addon export request"),
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

	addonExport := simc.Parse(request.SimcAddonExport)
	if !simc.HasRecognizedData(addonExport) {
		ginContext.JSON(http.StatusBadRequest, api_types.ErrorResponse{
			Message: utils.StrPtr("no recognizable addon export data found"),
		})

		return
	}

	ginContext.JSON(http.StatusOK, api_types.ParseAddonExportResponse{
		AddonExport: addonExport,
	})
}
