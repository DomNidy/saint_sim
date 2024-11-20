package resultlistener

import (
	"context"
	"fmt"
	"log"

	"github.com/DomNidy/saint_sim/apps/discord_bot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// stores discord user id and discord channel id of user who invoked a simrequest
type SimRequestOrigin struct {
	DiscordUserId    string
	DiscordChannelId string
}

// Maps an outgoing simulation request id back to the request origin
var OutboundSimRequests = make(map[string]*SimRequestOrigin)

// We will listen for new sim result trigger to be executed
// This is so we can respond to discord users with the sim results
func ListenForSimResults(ctx context.Context, conn *pgxpool.Conn, s *discordgo.Session) error {

	_, err := conn.Exec(ctx, "listen new_simulation_data")
	if err != nil {
		log.Fatalf("Failed to listen on new_simulation_data channel:")
	}

	log.Printf("listening for new sim data...")
	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			log.Printf("Error while waiting for notification: %v", err)
			continue
		}
		log.Printf("new simulation_data received, id: %v", notification.Payload)

		// Create struct to store simulation_data in
		simRes := struct {
			simDataId int
			simReqId  string
			data      string
		}{}
		// Query for simulation_data and scan it into the struct
		err = conn.QueryRow(ctx, "select id, request_id, sim_result from simulation_data where request_id = $1", notification.Payload).Scan(&simRes.simDataId, &simRes.simReqId, &simRes.data)
		if err != nil {
			log.Printf("Error while scanning to simRes: %v", err)
			continue
		}

		// Find the user who requested this sim
		requestOrigin, exists := OutboundSimRequests[simRes.simReqId]
		if !exists {
			log.Printf("Received notification of sim data, but no mapping to the request origin exists.")
			continue
		}

		// Send the simulation_data in the channel the sim was requested from
		// TODO: We Need to perform transformations on the simulation_data as it's excessively long
		// TODO: we should not need to truncate it to 2000 chars, but this is the discord bot limit
		// TODO: Also, prob should remove the mapping from the map after it gets consumed here
		messageContent := utils.ParseSimcReport(simRes.data, fmt.Sprintf("<@%v>, your sim request %v has been processed:\n", requestOrigin.DiscordUserId, simRes.simReqId))
		_, err = s.ChannelMessageSend(requestOrigin.DiscordChannelId, messageContent)
		if err != nil {
			log.Printf("Failed to send message in discord channel with sim data: %v", err)
			continue
		}
	}

}

// Adds a mapping from a simulation request id, to the discord user who
// request it. When we get notified that the simulation completed,
// we can use the mapping to notify the discord user of the results.
func AddOutboundSimRequestMapping(simRequestId string, requestOrigin *SimRequestOrigin) {
	OutboundSimRequests[simRequestId] = requestOrigin
	log.Printf("Added outbound sim request mapping: %v (sim req id) -> %v (discord user id)", simRequestId, requestOrigin.DiscordUserId)
}
