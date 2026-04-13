package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func commandHandlers() map[SaintCommandInteraction]func(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) error {
	return map[SaintCommandInteraction]func(
		session *discordgo.Session,
		interaction *discordgo.InteractionCreate,
	) error{
		SaintHelp: handleInteractionSaintHelp,
	}
}

// handleInteractionSaintHelp responds with the help message.
func handleInteractionSaintHelp(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) error {
	interResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{ //nolint:exhaustruct
			Content: "This is the help page.",
		},
	}

	err := session.InteractionRespond(interaction.Interaction, interResponse)
	if err != nil {
		return fmt.Errorf("respond to help interaction: %w", err)
	}

	return nil
}
