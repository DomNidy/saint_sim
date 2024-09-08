package constants

import "github.com/DomNidy/saint_sim/pkg/secrets"

// Load necessary env vars for the bot
// If we fail to load a secret here, we will panic
var (
	DiscordToken  = secrets.LoadSecret("DISCORD_TOKEN")
	ApplicationID = secrets.LoadSecret("APPLICATION_ID")
	// todo: find a better way of doing this
	GuildID = "897919702565814323"
)
