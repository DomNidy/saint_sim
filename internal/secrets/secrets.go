package secrets

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	DiscordToken string
	DBUser       string
	DBPassword   string
	DBHost       string
)

// The init() method is ran automatically when this package is imported
func init() {
	fmt.Println("Loading secrets")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	DiscordToken = getEnv("DISCORD_TOKEN", "")
	DBUser = getEnv("DB_USER", "")
	DBPassword = getEnv("DB_PASSWORD", "")
	DBHost = getEnv("DB_HOST", "")

	fmt.Printf("Loaded DiscordToken: %s\n", maskToken(DiscordToken, 3))
	fmt.Printf("Loaded DBUser: %s\n", maskToken(DBUser, 3))
	fmt.Printf("Loaded DBPassword: %s\n", maskToken(DBPassword, 3))
	fmt.Printf("Loaded DBHost: %s\n", maskToken(DBHost, 3))
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Used to print out the secrets to console
func maskToken(token string, visibleChars int) string {
	if len(token) < 3 {
		return token
	}
	return token[:visibleChars] + "XXX"
}
