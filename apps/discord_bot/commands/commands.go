package commands

import (
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

var Commands = []discordgo.ApplicationCommand{
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
		Name: string(SaintHelp),

		Description: "View help",
	},
}
