package api_utils

import (
	"fmt"
	"log"
	"net/http"

	"crypto/sha256"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	uuid "github.com/google/uuid"
)

// Use to generate UUID for simulation operations & results
func GenerateUUID() string {
	return uuid.New().String()
}

func HashApiKey(apiKey string) string {
	bytes := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", bytes)
}

// Check to see if a WoWCharacter actually exists on wow armory
func CheckWowCharacterExists(character *interfaces.WowCharacter) (bool, error) {
	url := fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-us/character/%s/%s/%s", character.Region, character.Realm, character.CharacterName)
	log.Printf("Checking if char exists at url: %v", url)
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 200 {
		return false, nil
	}

	return true, nil
}
