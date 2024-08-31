package main

import (
	"fmt"

	"github.com/DomNidy/sim-bot/internal/secrets"
)

func main() {
	// Remove this, just putting this here to shut the go linter up
	fmt.Println(secrets.DBHost, secrets.DBPassword, secrets.DBUser, secrets.DiscordToken)
}
