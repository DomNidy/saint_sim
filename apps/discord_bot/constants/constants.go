package constants

import "github.com/DomNidy/saint_sim/pkg/secrets"

// Load necessary env vars for the bot
// If we fail to load a secret here, we will panic
var (
	DBHost       = secrets.LoadSecret("DB_HOST")
	DBUser       = secrets.LoadSecret("DB_USER")
	DBPassword   = secrets.LoadSecret("DB_PASSWORD")
	DiscordToken = secrets.LoadSecret("DISCORD_TOKEN")
)
