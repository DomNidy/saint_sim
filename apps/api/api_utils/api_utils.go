package api_utils

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"crypto/sha256"

	uuid "github.com/google/uuid"

	api_types "github.com/DomNidy/saint_sim/pkg/api_types"
)

var (
	errNilWowCharacterProvided = errors.New("no wow character provided")

	// ErrUnexpectedStatusCodeReceivedFromArmory that blizzard armory returned a non-200 status
	// code.
	// We interpret 200 status code as "this character DOES exist".
	ErrUnexpectedStatusCodeReceivedFromArmory = errors.New(
		"received non 200 status code from blizzard armory",
	)

	// ErrCharacterNotExistsOnArmory is returned when a wow character could not
	// be found on the blizzard armory.
	ErrCharacterNotExistsOnArmory = errors.New("character not found on armory")
)

// GenerateUUID generates a UUID and returns it as a string.
func GenerateUUID() string {
	return uuid.New().String()
}

// HashAPIKey returns a SHA256 hash of an API key string.
//
// You should hash only the secret portion of the string;
// the stored secret hash in the db excludes the prefix.
// i.e., remove the "sk_xxx_" prefix, then hash the
// resulting string.
func HashAPIKey(apiKey string) string {
	bytes := sha256.Sum256([]byte(apiKey))

	return hex.EncodeToString(bytes[:])
}

// CheckWowCharacterExists verifies that a WoW character exists by
// sending an HTTP GET request to the blizzard armory. Returns an
// error if character does not exist, request fails, or unexpected
// status code returned.
func CheckWowCharacterExists(
	ctx context.Context,
	client *http.Client,
	character *api_types.WowCharacter,
) error {
	if character == nil {
		return errNilWowCharacterProvided
	}

	url := fmt.Sprintf(
		"https://worldofwarcraft.blizzard.com/en-us/character/%s/%s/%s",
		character.Region,
		character.Realm,
		character.CharacterName,
	)
	log.Printf("Checking if char exists at url: %v", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w: error making request", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: error checking character", err)
	}

	defer func() {
		// Q: Why do we need to read the entire response body?
		// A: We want to reuse http connection (HTTP/1.1 Keep-Alive header)
		// to reduce TCP & TLS handshake overhead. A HTTP client can only
		// reuse a connection for multiple requests if it knows where the
		// response body ends; underlying TCP connection is just a stream
		// of bytes.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf(
				"%w: could not find character '%v' on armory",
				ErrCharacterNotExistsOnArmory,
				character.CharacterName,
			)
		}

		return fmt.Errorf(
			"%w: got status code of %v",
			ErrUnexpectedStatusCodeReceivedFromArmory,
			resp.StatusCode,
		)
	}

	return nil
}
