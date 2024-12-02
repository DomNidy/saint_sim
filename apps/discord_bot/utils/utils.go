package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/DomNidy/saint_sim/pkg/interfaces"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	"github.com/bwmarrin/discordgo"
)

var log = logging.GetLogger()

// Pass the slice of interaction options received from discord command here
func ValidateInteractionSimOptions(appCmdInteractionData []*discordgo.ApplicationCommandInteractionDataOption) (*interfaces.SimulationOptions, error) {
	simOptions := interfaces.SimulationOptions{
		WowCharacter: interfaces.WowCharacter{},
	}

	for _, option := range appCmdInteractionData {
		switch option.Name {
		case "character_name":
			if characterName, ok := option.Value.(string); ok {
				simOptions.WowCharacter.CharacterName = characterName
			} else {
				return nil, fmt.Errorf("character_name must be a string")
			}
		case "realm":
			if realm, ok := option.Value.(string); ok {
				simOptions.WowCharacter.Realm = interfaces.WowCharacterRealm(realm)
			} else {
				return nil, fmt.Errorf("realm must be a string")
			}
		case "region":
			if region, ok := option.Value.(string); ok {
				simOptions.WowCharacter.Region = interfaces.WowCharacterRegion(region)
			} else {
				return nil, fmt.Errorf("region must be a string")
			}
		}
	}

	// make sure all fields are defined and not empty strings
	if simOptions.WowCharacter.CharacterName == "" {
		return nil, fmt.Errorf("character_name is missing or empty")
	}
	if simOptions.WowCharacter.Realm == "" {
		return nil, fmt.Errorf("realm is missing or empty")
	}
	if simOptions.WowCharacter.Region == "" {
		return nil, fmt.Errorf("region is missing or empty")
	}

	if !saintutils.IsValidWowRealm(string(simOptions.WowCharacter.Realm)) {
		return nil, fmt.Errorf("invalid wow realm")
	}
	if !saintutils.IsValidWowRegion(string(simOptions.WowCharacter.Region)) {
		return nil, fmt.Errorf("invalid wow reigon")
	}

	if !saintutils.IsValidSimOptions(&simOptions) {
		return nil, fmt.Errorf("invalid sim options according to saintutils")
	}

	return &simOptions, nil
}

func SendSimulationRequest(s *discordgo.Session, i *discordgo.InteractionCreate, options *interfaces.SimulationOptions) (*interfaces.SimulationResponse, error) {
	url := constants.SaintGatewayUrl.Value() + "/simulate"
	jsonData, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	// Send the sim request to API
	apiReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request %v", err)
	}
	apiReq.Header.Set("Api-Key", constants.SaintGatewayKey.Value())

	// TODO: Probably should stop creating this on each call of this function (http.Client caches tcp connections in internal state)
	// TODO: we can just re-use the client, and it's concurrency safe for multiple goroutines
	client := &http.Client{}
	resp, err := client.Do(apiReq)
	if err != nil {
		log.Errorf("error sending request to api: %v", err)
		return nil, fmt.Errorf("internal server error occurred")
	}
	defer resp.Body.Close()

	// this only occurs when the discord bot fails to authenticate with it's api key
	if resp.StatusCode == http.StatusForbidden {
		log.Warnf("Failed to authenticate with saint API. Please ensure the API key has been set in the environment variables, and is correct.")
		return nil, fmt.Errorf("internal server error occured, please try again later")
	}

	// TODO: Currently, a lot of our api responses dont actually correctly use the ErrorResponse interface,
	// TODO: so this interface assertions is never correct.
	// TODO: we should update the api responses to use that interface.
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
		log.Errorf("failed to unmarshal json response data: %v", err)
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

// todo: deprecate this and modify the simulation_worker to handling parsing of the results
// todo: this will work for now though
func ParseSimcReport(data, mentionUser string) string {
	reg, err := regexp.Compile(`([D|H]PS *\w+:(\n *[0-9]+\b .*)+|https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)|((\bConstant\b Buffs:)\n *(\b.*\b)*))`)
	if err != nil {
		log.Warnf("Failed to compile regular expression, simply returning the truncated sim data")
		return data[0:1000]
	}

	matches := reg.FindAll([]byte(data), -1)
	var sb strings.Builder
	sb.WriteString(mentionUser + "\n")
	// todo: since the regex is scuffed and captures buffs group twice,
	// todo: i am omitting the iteration over the last match with -1
	for _, match := range matches[0 : len(matches)-1] {
		sb.WriteString("\n--\n" + string(match))
	}
	final := sb.String()

	if len(final) > 1000 {
		return final[:1000]
	}
	return final
}
