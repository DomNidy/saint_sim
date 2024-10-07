package commands

import (
	"fmt"

	resultlistener "github.com/DomNidy/saint_sim/apps/discord_bot/result_listener"
	utils "github.com/DomNidy/saint_sim/apps/discord_bot/utils"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"

	"github.com/bwmarrin/discordgo"
)

var CommandHandlers = map[SaintCommandInteraction]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
	SaintSimulate: handleInteraction_SaintSimulate,
	SaintHelp:     handleInteraction_SaintHelp,
}

// Checks that the data struct of the Interaction object is of type `ApplicationCommandInteractionData`
func validateApplicationCommandInteractionData(i *discordgo.InteractionCreate) (*discordgo.ApplicationCommandInteractionData, error) {
	data, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData)

	if ok {
		return &data, nil
	}
	return nil, fmt.Errorf("received incorrect interaction type")
}

// Handlers simulation request from a discord user, and forwards their request to saint api
func handleInteraction_SaintSimulate(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Ensure the received interaction has data of type ApplicationCommandInteractionData
	interactionData, err := validateApplicationCommandInteractionData(i)

	// Handle case where this handler receives incorrect interaction type (we need application command interactions only)
	if err != nil {
		interResponse := utils.CreateErrorInteractionResponse("Something went wrong, please try again")
		s.InteractionRespond(i.Interaction, &interResponse)
		return err
	}

	// Validate the sim options
	simOptions, err := utils.ValidateInteractionSimOptions(interactionData.Options)
	if err != nil {
		interResponse := utils.CreateErrorInteractionResponse("Invalid arguments, please try again")
		s.InteractionRespond(i.Interaction, &interResponse)
		return err
	}

	// Send simulation request to api
	simRes, err := utils.SendSimulationRequest(s, i, simOptions)
	if err != nil {
		interResponse := utils.CreateErrorInteractionResponse(fmt.Sprintf("Failed to create simulation request: %v", err.Error()))
		s.InteractionRespond(i.Interaction, &interResponse)
		return err
	}

	// Extract the user id from the interaction
	var userId string
	if member := i.Member; member != nil {
		userId = member.User.ID
	} else if discordUser := i.User; discordUser != nil {
		userId = discordUser.ID
	}
	// Create sim request id -> discord user id mapping
	requestOrigin := &resultlistener.SimRequestOrigin{
		DiscordUserId:    userId,
		DiscordChannelId: i.ChannelID,
	}
	resultlistener.AddOutboundSimRequestMapping(*simRes.SimulationRequestId, requestOrigin)

	// Create discord response object
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: *saintutils.StrPtr(fmt.Sprintf("Simulation was successfully submitted. (operation_id: %s)", *simRes.SimulationRequestId)),
		},
	}
	// Send discord response, indicating status of their simulation request
	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		return err
	}

	return nil
}

// Responds with help message
func handleInteraction_SaintHelp(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	interResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: *saintutils.StrPtr("This is the help page\n`/simulate`: Perform sim request for your WoW character."),
		},
	}

	err := s.InteractionRespond(i.Interaction, interResponse)
	if err != nil {
		return err
	}

	return nil
}
