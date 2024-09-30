package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DomNidy/saint_sim/apps/discord_bot/commands"
	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/DomNidy/saint_sim/apps/discord_bot/utils"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var s *discordgo.Session
var db *pgxpool.Pool

// Utility function used to create an erroneous discord response message
// (a message that indicates something went wrong)
func createErrorInteractionResponse(msg string) discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}
}

// When we receive a simulation request and forward it to the api, spawn a new thread and periodically
// check for results for the simulation with this sim id
type SimulationInteractionContext struct {
	DiscordMessageId string // the discord message id that the sim was initiated from
	SimulationId     string // the resulting sim id
}

var (
	commandHandlers = map[commands.SaintCommandInteraction]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		commands.SaintSimulate: func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			// Handle case where this handler receives incorrect interaction type (we need application command interactions only)
			if _, ok := i.Interaction.Data.(discordgo.ApplicationCommandInteractionData); !ok {
				errResponse := createErrorInteractionResponse("Something went wrong, please try again")
				err := s.InteractionRespond(i.Interaction, &errResponse)
				if err != nil {
					log.Panicf("error sending error response: %v\n", errResponse.Data)
				}
				return
			}

			simOptions, err := utils.ValidateInteractionSimOptions(i.ApplicationCommandData().Options)
			if err != nil {
				log.Printf("invalid sim options received: %v", simOptions)
				errResponse := createErrorInteractionResponse("Invalid arguments, please try again.")
				err := s.InteractionRespond(i.Interaction, &errResponse)
				if err != nil {
					log.Printf("Error sending error response: %v", err)
				}
				return
			}

			// Send simulation request to api
			iJson, err := json.Marshal(i.Interaction)
			if err != nil {
				log.Printf("failed to marshal interaction to json: %v", err)
			}
			log.Printf("%v", string(iJson))

			simRes, err := utils.SendSimulationRequest(s, i, simOptions)
			if err != nil {
				errResponse := createErrorInteractionResponse("Failed to create simulation request")
				err := s.InteractionRespond(i.Interaction, &errResponse)
				log.Printf("Failed to create simulation request: %v", err)
				return
			}

			// Create discord response object
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: *saintutils.StrPtr(fmt.Sprintf("Simulation was successfully submitted. (operation_id: %s)", *simRes.SimulationId)),
				},
			}

			// Send discord response, indicating status of their simulation request
			err = s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println(response.Data)
				panic(err)
			}

			// TODO: i.Message.ID is undefined because the user doesnt actually send a normal message, its an interaction
			// TODO: inspect the json printed in the logs to figure out how to provide status updates (maybe just send our own messages?)
			// TODO: Maybe use defer?
			// If all was good, spawn up new thread to periodically check for sim results

			log.Printf("inter type: %v", i.Member.User.ID)
		},
	}
)

// We will listen for new sim result trigger to be executed
// This is so we can respond to discord users with the sim results
func ListenForSimResults(ctx context.Context, conn *pgxpool.Conn) error {

	_, err := conn.Exec(ctx, "listen new_simulation_data")
	if err != nil {
		log.Fatalf("Failed to listen on new_simulation_data channel:")
	}

	log.Printf("listening for new sim data...")
	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		log.Printf("notification received: %v", notification.Payload)
	}

}

func init() {
	// Setup discord bot
	fmt.Println("Loaded secrets:")
	fmt.Printf("%s: %s\n", constants.DiscordToken.Key(), constants.DiscordToken.MaskedValue())

	var err error
	s, err = discordgo.New("Bot " + constants.DiscordToken.Value())
	if err != nil {
		log.Fatalf("Error occured during discord session creation: %v", err)
		return
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Add interaction handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			fmt.Printf("Interaction occurred: %v\n", i.ApplicationCommandData().Name)
			if h, ok := commandHandlers[commands.SaintCommandInteraction(i.ApplicationCommandData().Name)]; ok {
				h(s, i)
			}
		default:
			fmt.Printf("Received interaction of type %v, but we do not have any handlers for this type of interaction", i.Type)
		}
	})

	// Register application commands
	log.Printf("Registering commands...")

	cmdIDS := make(map[string]string, len(commands.Commands))
	for _, cmd := range commands.Commands {
		rcmd, err := s.ApplicationCommandCreate(constants.ApplicationID.Value(), "", &cmd)
		if err != nil {
			log.Fatalf("Failed to register command with name '%q': %v", cmd.Name, err)
		} else {
			fmt.Printf("Registered command %q\n", cmd.Name)
		}

		cmdIDS[rcmd.ID] = rcmd.Name
	}
}

func main() {
	ctx := context.Background()
	// Setup postgres connection
	db := saintutils.InitPostgresConnectionPool(ctx)
	defer db.Close()

	// get a connection to listen for sim result trigger
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Fatalf("Failed to get conn from pool: %v", err)
	}

	go ListenForSimResults(ctx, conn)

	// Open websocket connection
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
