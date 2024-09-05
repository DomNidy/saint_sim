package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
	"github.com/bwmarrin/discordgo"
)

func main() {
	fmt.Println("Loaded secrets:")
	fmt.Printf("%s: %s\n", constants.DBHost.Key(), constants.DBHost.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBUser.Key(), constants.DBUser.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBPassword.Key(), constants.DBPassword.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DiscordToken.Key(), constants.DiscordToken.MaskedValue())

	discord, err := discordgo.New("Bot " + constants.DiscordToken.Value())
	if err != nil {
		log.Fatalf("Error occured during discord session creation: %v", err)
		return
	}
	// Add event listener to respond to commmands
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if m.Content == "!hello" {
			s.ChannelMessageSend(m.ChannelID, "Hello, "+m.Author.Username+"!")
		}
	})

	// Open websocket connection to discord
	err = discord.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer discord.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
