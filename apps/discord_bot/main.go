package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DomNidy/saint_sim/apps/discord_bot/commands"
	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var s *discordgo.Session
var db *pgxpool.Pool

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

	// Register application commands
	commands.RegisterApplicationCommands(s, "")

	// Add handlers to session so the bot can respond to events
	commands.AddHandlers(s)
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
