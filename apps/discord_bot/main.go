package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/DomNidy/saint_sim/pkg/utils"
	"github.com/bwmarrin/discordgo"
)

// Slash commands

const (
	SaintCommandSimulate string = "simulate"
	SaintCommandHelp     string = "help"
)

var (

	// Slash commands
	commands = []discordgo.ApplicationCommand{
		{
			Name:        "simulate",
			Description: "Simulate your characters DPS.",
		},
		{
			Name:        "modals-survey",
			Description: "Take a survey about modals",
		},
	}

	// componentHandlers deal with responding to interactions from message components
	// for example, when a user selects a field, an interaction of type MessageComponent occurs
	// but when a user enters a slash command, an interaction of type ApplicationCommandData will be occur
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"simulate_character_realm": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "5 DPS",
				},
			})

			if err != nil {
				panic(err)
			}
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// https://github.com/kevcenteno/discordgo/blob/f8c5d6c837ef0cd4db6a4b7d03e301d83f3708c4/examples/components/main.go
		"simulate": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Simulate your WoW characters' DPS",
					Flags:   discordgo.MessageFlagsEphemeral,

					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID:    "simulate_character_realm",
									Placeholder: "Select Character Realm",
									MinValues:   utils.IntPtr(1),
									MaxValues:   1,
									Options: []discordgo.SelectMenuOption{
										{
											Label:       "Hydraxis",
											Value:       "hydraxis",
											Description: "Are you on Hydraxis?",
										},
										{
											Label:       "Area-52",
											Value:       "area-52",
											Description: "Are you on Area-52?",
										},
									},
								},
							},
						},
					},
				}}

			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				panic(err)
			}
		},
	}
)

var s *discordgo.Session

// Create session object
func init() {
	fmt.Println("Loaded secrets:")
	fmt.Printf("%s: %s\n", constants.DBHost.Key(), constants.DBHost.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBUser.Key(), constants.DBUser.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBPassword.Key(), constants.DBPassword.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DiscordToken.Key(), constants.DiscordToken.MaskedValue())

	var err error
	s, err = discordgo.New("Bot " + constants.DiscordToken.Value())
	if err != nil {
		log.Fatalf("Error occured during discord session creation: %v", err)
		return
	}
}

func init() {

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Add interaction handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			fmt.Printf("Interaction occurred: %v\n", i.ApplicationCommandData().Name)
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			fmt.Printf("Interaction occurred: %v\n", i.MessageComponentData().CustomID)
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		default:
			fmt.Printf("Received interaction of type %v, but we do not have any handlers for this type of interaction", i.Type)
		}

	})

	// Register application commands
	log.Printf("Registering commands...")

	cmdIDS := make(map[string]string, len(commands))
	for _, cmd := range commands {
		rcmd, err := s.ApplicationCommandCreate(constants.ApplicationID.Value(), constants.GuildID, &cmd)
		if err != nil {
			log.Fatalf("Failed to register command with name '%q': %v", cmd.Name, err)
		} else {
			fmt.Printf("Registered command %q\n", cmd.Name)
		}

		cmdIDS[rcmd.ID] = rcmd.Name
	}
}

func main() {

	// Open websocket connection
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
