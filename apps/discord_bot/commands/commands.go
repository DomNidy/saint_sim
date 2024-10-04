package commands

import (
	"log"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/DomNidy/saint_sim/pkg/utils"

	"github.com/bwmarrin/discordgo"
)

// Interactions fired off in response to application commands (slash commands)
type SaintCommandInteraction string

const (
	SaintSimulate SaintCommandInteraction = "simulate"
	SaintHelp     SaintCommandInteraction = "help"
	// Slash commands
)

var ApplicationCommands = []discordgo.ApplicationCommand{
	{
		Name:        string(SaintSimulate),
		Description: "Simulate your characters DPS.",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Description: "What region do you play on?",
				Name:        "region",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "eu",
						Value: "eu",
					},
					{
						Name:  "us",
						Value: "us",
					},
					{
						Name:  "kr",
						Value: "kr",
					}, {
						Name:  "tw",
						Value: "tw",
					}, {
						Name:  "cn",
						Value: "cn",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Name:        "realm",
				Description: "What realm is your character on?",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "us-thrall",
						Value: "thrall",
					},
					{
						Name:  "us-hydraxis",
						Value: "hydraxis",
					},
					{
						Name:  "eu-silvermoon",
						Value: "silvermoon",
					}, {
						Name:  "eu-draenor",
						Value: "draenor",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Description: "What is your characters name?",
				Name:        "character_name",
				MinLength:   utils.IntPtr(2),
				MaxLength:   12,
			},
		},
	},
	{
		Name:        string(SaintHelp),
		Description: "View help",
	},
}

// Registers all ApplicationCommands defined above
// if guildId is empty, we will register this command
// globally, (any server with the bot will have the commands)
func RegisterApplicationCommands(s *discordgo.Session, guildId string) {
	for _, appCommand := range ApplicationCommands {
		_, err := s.ApplicationCommandCreate(constants.ApplicationID.Value(), guildId, &appCommand)
		if err != nil {
			log.Fatalf("Failed to register command with name '%q': %v", appCommand.Name, err)
		} else {
			log.Printf("Registered command %q\n", appCommand.Name)
		}

	}
}

// Adds the necessary handlers to the bot session
// we add a handler to be notified when session is ready,
// and a handler to be notified of interactions
func AddHandlers(s *discordgo.Session) {

	// Let us know when the session is ready
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Receives interactions and 'routes' them to the associated interaction handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			cmdName := SaintCommandInteraction(i.ApplicationCommandData().Name)
			h, ok := CommandHandlers[cmdName]
			// ensure that handler with matching name exists
			if !ok {
				log.Printf("Failed to find command handler for command '%v'", cmdName)
				return
			}
			// execute command handler
			err := h(s, i)
			if err != nil {
				log.Printf("Error occured while executing application command handler: %v", err)
			}
		default:
			log.Printf("Received interaction of type %v, but we do not have any handlers for this type of interaction", i.Type)
		}
	})
}
