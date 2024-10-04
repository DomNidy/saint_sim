package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
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

func SendSimulationRequest(s *discordgo.Session, i *discordgo.InteractionCreate, options *interfaces.SimulationOptions) (*interfaces.SimulationResponse, error) {
	url := constants.SaintApiUrl.Value() + "/simulate"
	jsonData, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	// Send the sim request to API
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error sending request to api: %v", err)
		return nil, fmt.Errorf("internal server error occurred")
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		var errResp interface{}
		decodeErr := json.NewDecoder(resp.Body).Decode(&errResp)
		// Ensure the returned type correctly matches ErrorResponse type
		apiErr, ok := errResp.(interfaces.ErrorResponse)
		if !ok || decodeErr != nil {
			return nil, fmt.Errorf("could not find WoW character")
		} else if apiErr.Message == nil {
			// Ensure apiErr.Message is not nil before dereferencing it
			return nil, fmt.Errorf("could not find WoW character")
		}

		return nil, fmt.Errorf(*apiErr.Message)
	}

	var simRespose interfaces.SimulationResponse
	// Strict decoder
	// this will return an error if an unknown field is returned from the response json
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&simRespose)
	if err != nil {
		log.Printf("failed to unmarshal json response data: %v", err)
		return nil, fmt.Errorf("internal server error occured")
	}

	return &simRespose, nil
}

// Utility function used to create an erroneous discord response message
// (a message that indicates something went wrong)
func CreateErrorInteractionResponse(msg string) discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}
}
