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
	resultlistener "github.com/DomNidy/saint_sim/apps/discord_bot/result_listener"
	"github.com/DomNidy/saint_sim/internal/platform/postgres"
	"github.com/DomNidy/saint_sim/internal/secrets"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

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
	pool, err := postgres.New(context.Background(), postgres.Credentials{
		DBUser:     secrets.LoadSecret("DB_USER").Value(),
		DBPassword: secrets.LoadSecret("DB_PASSWORD").Value(),
		DBHost:     secrets.LoadSecret("DB_HOST").Value(),
		DBName:     secrets.LoadSecret("DB_NAME").Value(),
		DBPort:     "5432",
	})

	if err != nil {
		log.Panicf("%s: could not create postgres pool", err)
	}
	defer pool.Close()

	// get a connection to listen for sim result trigger
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Panicf("Failed to get conn from pool: %v", err)
	}

	go resultlistener.ListenForSimResults(ctx, conn, s)

	// Open websocket connection
	err = s.Open()
	if err != nil {
		log.Panicf("Cannot open the session: %v", err)
	}
	defer s.Close()

	log.Printf("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
