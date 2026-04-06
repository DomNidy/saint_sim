package utils

import (
	"regexp"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
)

// IsValidSimOptions validates the user-provided simulation options.
func IsValidSimOptions(options *api_types.SimulationOptions) bool {
	return isValidInput(options.WowCharacter.CharacterName) &&
		isValidInput(string(options.WowCharacter.Realm)) &&
		isValidInput(string(options.WowCharacter.Region))
}

func isValidInput(input string) bool {
	valid := regexp.MustCompilePOSIX(`^[[:alnum:]_-]+$`)

	return valid.MatchString(input)
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
