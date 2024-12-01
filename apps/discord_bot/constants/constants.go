package constants

import "github.com/DomNidy/saint_sim/pkg/secrets"

// Load necessary env vars for the bot
// If we fail to load a secret here, we will panic
var (
	DiscordToken    = secrets.LoadSecretFromEnv("DISCORD_TOKEN")
	ApplicationID   = secrets.LoadSecretFromEnv("APPLICATION_ID")
	SaintGatewayUrl = secrets.LoadSecretFromEnv("SAINT_GATEWAY_URL")
	SaintGatewayKey = secrets.LoadSecretFromEnv("SAINT_GATEWAY_KEY")
	// todo: find a better way of doing this
	GuildID = "640276404474347520"
)
