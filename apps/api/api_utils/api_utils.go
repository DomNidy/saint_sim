package api_utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/pkg/interfaces"
	uuid "github.com/google/uuid"
)

// Use to generate UUID for simulation operations & results
func GenerateUUID() string {
	return uuid.New().String()
}

// Check to see if a WoWCharacter actually exists on wow armory
func CheckWowCharacterExists(character *interfaces.WowCharacter) (bool, error) {
	url := fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-us/character/%v/%v/%v", character.Region, character.Realm, character.CharacterName)
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
