package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DomNidy/saint_sim/apps/discord_bot/commands"
	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	resultlistener "github.com/DomNidy/saint_sim/apps/discord_bot/result_listener"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

var log = logging.GetLogger()

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

	go resultlistener.ListenForSimResults(ctx, conn, s)

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
