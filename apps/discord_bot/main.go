package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DomNidy/saint_sim/apps/discord_bot/commands"
	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/DomNidy/saint_sim/pkg/interfaces"
	"github.com/DomNidy/saint_sim/pkg/utils"
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

func periodicCheckSimStatus(ctx context.Context, simInteraction SimulationInteractionContext) {
	log.Printf("Periodically checking sim %v, for msg id %v", simInteraction.SimulationId, simInteraction.DiscordMessageId)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Periodic check sim jod %v cancelled", simInteraction.SimulationId)
			return
		default:
			time.Sleep(2 * time.Second)
			log.Printf("Checking sim status for sim %v", simInteraction.SimulationId)

			var simRes string
			err := db.QueryRow(ctx, "select sim_result from simulation_data where from_request = $1", simInteraction.SimulationId).Scan(&simRes)
			if err != nil {
				log.Printf("didnt find sim result yet")
				continue
			}

			log.Printf("%v", simRes)
			return

		}
	}
}

var (
	commandHandlers = map[commands.SaintCommandInteraction]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// https://github.com/bwmarrin/discordgo/tree/master/examples/components
		// https://github.com/kevcenteno/discordgo/blob/f8c5d6c837ef0cd4db6a4b7d03e301d83f3708c4/examples/components/main.go
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

			fmt.Printf("sim command data: %v\n", i.Interaction.ApplicationCommandData())
			// Unmarshall received options into SimulationOptions struct so we can validate it
			// seems like we need to explicitly create the WoWCharacter struct inside, as go wont allocate
			// memory for the struct itsself (because we accept a pointer to a struct), so it just
			// allocates memory for the pointer, not the struct.
			simOptions := interfaces.SimulationOptions{
				WowCharacter: &interfaces.WoWCharacter{},
			}

			// todo: clean this bad parsing up
			for _, option := range i.ApplicationCommandData().Options {
				switch option.Name {
				case "character_name":
					if characterName, ok := option.Value.(string); ok {
						simOptions.WowCharacter.CharacterName = &characterName
					}
				case "realm":
					if realm, ok := option.Value.(string); ok {
						simOptions.WowCharacter.Realm = &realm
					}
				case "region":
					if region, ok := option.Value.(string); ok {
						simOptions.WowCharacter.Region = &region
					}

				}
			}

			if !utils.IsValidSimOptions(&simOptions) {
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

			simRes, err := sendSimulationRequest(s, i, &simOptions)
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
					Content: *utils.StrPtr(fmt.Sprintf("Simulation was successfully submitted. (operation_id: %s)", *simRes.SimulationId)),
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
			go periodicCheckSimStatus(context.Background(), SimulationInteractionContext{DiscordMessageId: i.Message.ID, SimulationId: *simRes.SimulationId})
		},
	}
)

func sendSimulationRequest(s *discordgo.Session, i *discordgo.InteractionCreate, options *interfaces.SimulationOptions) (*interfaces.SimulationResponse, error) {
	url := "http://saint_api:8080/simulate"
	jsonData, err := json.Marshal(options)
	if err != nil {
		fmt.Printf("Error marshaling request data: %v\n", err)
		return nil, err
	}

	// Send the sim request to API
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status from api: %v", resp.StatusCode)
	}

	var simRespose interfaces.SimulationResponse

	// Strict decoder
	// this will return an error if an unknown field is returned from the response json
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&simRespose)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json response data: %v", err)
	}

	return &simRespose, nil
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
	db := utils.InitPostgresConnectionPool(ctx)
	defer db.Close()

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
