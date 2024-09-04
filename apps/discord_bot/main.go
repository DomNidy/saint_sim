package main

import (
	"fmt"

	"github.com/DomNidy/saint_sim/apps/discord_bot/constants"
)

func main() {
	fmt.Println("Loaded secrets:")
	fmt.Printf("%s: %s\n", constants.DBHost.Key(), constants.DBHost.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBUser.Key(), constants.DBUser.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DBPassword.Key(), constants.DBPassword.MaskedValue())
	fmt.Printf("%s: %s\n", constants.DiscordToken.Key(), constants.DiscordToken.MaskedValue())
}
