package utils

import (
	"context"
	"fmt"

	"github.com/DomNidy/saint_sim/apps/gateway/middleware"
)

// Utility function to extract the authenticated claims/data from context
// The authenticated claims/data are set after the Authenticate middleware is run
func GetAuthenticatedData(ctx context.Context) (interface{}, error) {
	claims := ctx.Value(middleware.AuthedDataKey)

	if claims == nil {
		return nil, fmt.Errorf("no authenticated data found in context")
	}

	return claims, nil
}
