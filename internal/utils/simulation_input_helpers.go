// Package utils provides shared helpers for validation, queueing, and other
// application integration concerns.
package utils

import (
	"fmt"

	api "github.com/DomNidy/saint_sim/internal/api"
)

type simulationValidationError string

func (err simulationValidationError) Error() string {
	return string(err)
}

const (
	errSimulationOptionsRequired simulationValidationError = "simulation options are required"
	errMissingSimcAddonExport    simulationValidationError = "simc addon export is required"
)

// ValidateSimOptions validates the user-provided simulation options and reports the first failure.
func ValidateSimOptions(options *api.SimulationOptions) error {
	if options == nil {
		return errSimulationOptionsRequired
	}

	if options.SimcAddonExport == "" {
		return fmt.Errorf("%w", errMissingSimcAddonExport)
	}

	return nil
}
