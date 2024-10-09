package constants

import "github.com/DomNidy/saint_sim/pkg/secrets"

// Load necessary env vars for the bot
// If we fail to load a secret here, we will panic
var (
	DiscordToken  = secrets.LoadSecret("DISCORD_TOKEN")
	ApplicationID = secrets.LoadSecret("APPLICATION_ID")
	SaintApiUrl   = secrets.LoadSecret("SAINT_API_URL")
	SaintApiKey   = secrets.LoadSecret("SAINT_API_KEY")
	// todo: find a better way of doing this
	GuildID = "640276404474347520"
)
