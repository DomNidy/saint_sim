package handlers

import (
	"context"
	"strings"

	api "github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/simc"
	"github.com/DomNidy/saint_sim/internal/utils"
)

// ParseAddonExport parses a SimC addon export and returns structured data.
func (server *Server) ParseAddonExport(
	_ context.Context,
	request api.ParseAddonExportRequestObject,
) (api.ParseAddonExportResponseObject, error) {
	_ = server

	if request.Body == nil {
		return invalidParseAddonExportRequestResponse(), nil
	}

	normalizedExport := simc.NormalizeLineEndings(request.Body.SimcAddonExport)
	if strings.TrimSpace(normalizedExport) == "" {
		return missingParseAddonExportResponse(), nil
	}

	addonExport := simc.Parse(normalizedExport)
	if !simc.HasRecognizedData(addonExport) {
		return unrecognizedParseAddonExportResponse(), nil
	}

	return api.ParseAddonExport200JSONResponse{
		AddonExport: addonExport,
	}, nil
}

func invalidParseAddonExportRequestResponse() api.ParseAddonExport400JSONResponse {
	return api.ParseAddonExport400JSONResponse{
		BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
			Message: utils.StrPtr("Invalid parse addon export request"),
		},
	}
}

func missingParseAddonExportResponse() api.ParseAddonExport400JSONResponse {
	return api.ParseAddonExport400JSONResponse{
		BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
			Message: utils.StrPtr("simc_addon_export is required"),
		},
	}
}

func unrecognizedParseAddonExportResponse() api.ParseAddonExport400JSONResponse {
	return api.ParseAddonExport400JSONResponse{
		BadRequestErrorJSONResponse: api.BadRequestErrorJSONResponse{
			Message: utils.StrPtr("no recognizable addon export data found"),
		},
	}
}
