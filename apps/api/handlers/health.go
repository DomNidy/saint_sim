package handlers

import (
	"context"

	api "github.com/DomNidy/saint_sim/internal/api"
)

// Health returns the API health status.
func (server *Server) Health(
	_ context.Context,
	request api.HealthRequestObject,
) (api.HealthResponseObject, error) {
	_ = server
	_ = request

	return api.Health200JSONResponse{
		Status: "healthy",
	}, nil
}
