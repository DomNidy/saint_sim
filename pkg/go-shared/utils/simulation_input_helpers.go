// Package utils provides shared helpers for validation, queueing, and other
// application integration concerns.
package utils

import (
	"fmt"
	"regexp"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
)

type simulationValidationError string

func (err simulationValidationError) Error() string {
	return string(err)
}

const (
	errSimulationOptionsRequired simulationValidationError = "simulation options are required"
	errInvalidCharacterName      simulationValidationError = "invalid character name"
	errInvalidRealm              simulationValidationError = "invalid realm"
	errInvalidRegion             simulationValidationError = "invalid region"
	simInputPattern                                        = `^[[:alnum:]_-]+$`
)

// ValidateSimOptions validates the user-provided simulation options and reports the first failure.
func ValidateSimOptions(options *api_types.SimulationOptions) error {
	if options == nil {
		return errSimulationOptionsRequired
	}

	switch {
	case !isValidInput(options.WowCharacter.CharacterName):
		return fmt.Errorf("%w %q", errInvalidCharacterName, options.WowCharacter.CharacterName)
	case !isValidInput(string(options.WowCharacter.Realm)):
		return fmt.Errorf("%w %q", errInvalidRealm, options.WowCharacter.Realm)
	case !isValidInput(string(options.WowCharacter.Region)):
		return fmt.Errorf("%w %q", errInvalidRegion, options.WowCharacter.Region)
	default:
		return nil
	}
}

// IsValidSimOptions validates the user-provided simulation options.
func IsValidSimOptions(options *api_types.SimulationOptions) bool {
	return ValidateSimOptions(options) == nil
}

func isValidInput(input string) bool {
	matched, _ := regexp.MatchString(simInputPattern, input)

	return input != "" && matched
}

// IsValidWowRegion reports whether region is in the allowlist.
func IsValidWowRegion(region string) bool {
	switch api_types.WowRegion(region) {
	case api_types.Us, api_types.Eu, api_types.Tw, api_types.Cn, api_types.Kr:
		return true
	default:
		return false
	}
}

// IsValidWowRealm reports whether realm is in the allowlist.
func IsValidWowRealm(realm string) bool {
	switch api_types.WowRealm(realm) {
	case api_types.Draenor, api_types.Hydraxis, api_types.Silvermoon, api_types.Thrall:
		return true
	default:
		return false
	}
}
