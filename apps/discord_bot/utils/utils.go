package utils

import (
	"fmt"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"
	"github.com/bwmarrin/discordgo"
)

// Pass the slice of interaction options received from discord command here
func ValidateInteractionSimOptions(appCmdInteractionData []*discordgo.ApplicationCommandInteractionDataOption) (*interfaces.SimulationOptions, error) {
	simOptions := interfaces.SimulationOptions{
		WowCharacter: &interfaces.WoWCharacter{},
	}

	for _, option := range appCmdInteractionData {
		switch option.Name {
		case "character_name":
			if characterName, ok := option.Value.(string); ok {
				simOptions.WowCharacter.CharacterName = &characterName
			} else {
				return nil, fmt.Errorf("character_name must be a string")
			}
		case "realm":
			if realm, ok := option.Value.(string); ok {
				simOptions.WowCharacter.Realm = &realm
			} else {
				return nil, fmt.Errorf("realm must be a string")
			}
		case "region":
			if region, ok := option.Value.(string); ok {
				simOptions.WowCharacter.Region = &region
			} else {
				return nil, fmt.Errorf("region must be a string")
			}
		}
	}

	// make sure all fields are defined and not empty strings
	if simOptions.WowCharacter.CharacterName == nil || *simOptions.WowCharacter.CharacterName == "" {
		return nil, fmt.Errorf("character_name is missing or empty")
	}
	if simOptions.WowCharacter.Realm == nil || *simOptions.WowCharacter.Realm == "" {
		return nil, fmt.Errorf("realm is missing or empty")
	}
	if simOptions.WowCharacter.Region == nil || *simOptions.WowCharacter.Region == "" {
		return nil, fmt.Errorf("region is missing or empty")
	}

	if !saintutils.IsValidSimOptions(&simOptions) {
		return nil, fmt.Errorf("invalid sim options according to saintutils")
	}

	return &simOptions, nil
}
