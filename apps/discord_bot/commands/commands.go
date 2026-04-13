// Package commands registers and handles Discord slash commands.
package commands

import (
	"log"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"

	"github.com/bwmarrin/discordgo"
)

// SaintCommandInteraction identifies a Discord application command.
type SaintCommandInteraction string

const (
	// SaintHelp shows the help command response.
	SaintHelp SaintCommandInteraction = "help"
)

func applicationCommands() []discordgo.ApplicationCommand {
	return []discordgo.ApplicationCommand{
		{ //nolint:exhaustruct
			Name:        string(SaintHelp),
			Description: "View help",
		},
	}
}

// RegisterApplicationCommands registers all application commands for the bot.
// If guildID is empty, the command is registered globally.
func RegisterApplicationCommands(session *discordgo.Session, guildID string) {
	commands := applicationCommands()

	for idx := range commands {
		appCommand := commands[idx]

		_, err := session.ApplicationCommandCreate(
			constants.ApplicationID.Value(),
			guildID,
			&appCommand,
		)
		if err != nil {
			log.Fatalf("Failed to register command with name '%q': %v", appCommand.Name, err)
		}

		log.Printf("Registered command %q\n", appCommand.Name)
	}
}

// AddHandlers attaches the runtime handlers for the Discord bot session.
func AddHandlers(session *discordgo.Session) {
	// Let us know when the session is ready
	session.AddHandler(func(discordSession *discordgo.Session, _ *discordgo.Ready) {
		log.Printf(
			"Logged in as: %v#%v",
			discordSession.State.User.Username,
			discordSession.State.User.Discriminator,
		)
	})

	// Receives interactions and 'routes' them to the associated interaction handler
	session.AddHandler(
		func(
			discordSession *discordgo.Session,
			interaction *discordgo.InteractionCreate,
		) {
			handlers := commandHandlers()

			if interaction.Type == discordgo.InteractionApplicationCommand {
				commandName := SaintCommandInteraction(interaction.ApplicationCommandData().Name)

				handler, ok := handlers[commandName]
				if !ok {
					log.Printf("Failed to find command handler for command '%v'", commandName)

					return
				}

				err := handler(discordSession, interaction)
				if err != nil {
					log.Printf(
						"Error occurred while executing application command handler: %v",
						err,
					)
				}

				return
			}

			switch interaction.Type {
			case discordgo.InteractionApplicationCommand:
				return
			case discordgo.InteractionPing,
				discordgo.InteractionMessageComponent,
				discordgo.InteractionApplicationCommandAutocomplete,
				discordgo.InteractionModalSubmit:
				log.Printf(
					"Received interaction of type %v, but we do not have any handlers for this type of interaction",
					interaction.Type,
				)
			default:
				log.Printf(
					"Received interaction of type %v, but we do not have any handlers for this type of interaction",
					interaction.Type,
				)
			}
		},
	)
}
