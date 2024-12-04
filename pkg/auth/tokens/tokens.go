package tokens

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// String that identifies which front-end a token is valid for.
//
// A `RequestOrigin` can be extracted from an HTTP request. It is an identifier
// that details which front-end a request originated from (i.e., "web", or "discord_bot").
type RequestOrigin string

const (
	DiscordBotRequestOrigin RequestOrigin = "discord_bot"
	WebRequestOrigin        RequestOrigin = "web"
)

// Type for custom claim types.
//
// This type indicates that the concrete value is expected
// to be one of the claim struct types (e.g., `NativeUserClaims`)
// Better than just using `interface{}` type.
type UserClaims interface {
	// Return the claims as a map
	ToMap() (map[string]interface{}, error)
}

// JWT that native users use to authenticate.
// This is presented in a HTTP header as: "Authorization: Bearer <jwt>"
// Registered claim names: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
//
// This is distinct from the `ForeignUserClaims`
// used by the Saint Discord bot.
//
// A native user refers to an end user that directly
// interacts with the Saint API (i.e. their requests aren't
// proxied through the Saint Discord bot). This is in contrast
// to a foreign user.
type NativeUserClaims struct {
	jwt.RegisteredClaims `mapstructure:",squash"`
	// Used to identify the origin (Discord or web app) a token
	// is valid for.
	RequestOrigin RequestOrigin `json:"request_origin" mapstructure:",squash"`
}

// JWT that foreign users will be authenticated with.
// This is presented in a HTTP header as: "Authorization: Bot <jwt>"
// Registered claim names: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
//
// A foreign user refers to an end user that
// interacts with the Saint API through the
// Saint Discord bot. This is distinct from the
// `NativeUserClaims` used by native users.
//
// The reason there are two different JWT types is
// because the Saint Discord bot needs to authenticate
// different data.
//
// With this, we can make JWTs that do things like:
//   - discord-server-scoped JWTs (access control based
//     on the originating discord server)
//   - discord-user-scoped JWTs (access control based
//     on specific discord user id)

type ForeignUserClaims struct {
	jwt.RegisteredClaims `mapstructure:",squash"`
	DiscordClaims        `mapstructure:",squash"`

	// Used to identify the origin (Discord or web app) a token
	// is valid for.
	RequestOrigin RequestOrigin `json:"request_origin"`
}

// Custom JWT claims specific to requests originating from Discord
// These can be used to make a token server-scoped
type DiscordClaims struct {

	// Custom claims for Discord-specific context
	DiscordServerID *string `json:"discord_server_id,omitempty"` // The server ID for scoping access

	Permissions *[]string `json:"permissions,omitempty"` // Optional permissions granted to this token
}

// This method converts a `ForeignUserClaims` type to map[string]interface{}
//
// This is needed as the golang-jwt library expects the payload to be of type map[string]interface{}
// To perform this conversion, we simply marshal the struct into json, then unmarshal it
// into map[string]interface{} type. This method uses the struct tags of `ForeignUserClaims`
// to perform marshalling and unmarshalling.
func (t *ForeignUserClaims) ToMap() (map[string]interface{}, error) {

	res, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWT payload into json: %w", err)
	}

	var resMap map[string]interface{}
	if err := json.Unmarshal(res, &resMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT payload into map: %w", err)
	}

	return resMap, nil
}

// This method converts a `NativeUserClaims` type to map[string]interface{}
func (t *NativeUserClaims) ToMap() (map[string]interface{}, error) {
	res, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWT payload into json: %w", err)
	}

	var resMap map[string]interface{}
	if err := json.Unmarshal(res, &resMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT payload into map: %w", err)
	}

	return resMap, nil
}

// Creates a `ForeignUserClaims` object with the specified params.
func CreateDiscordUserClaims(discordUserID string, discordServerID *string, permissions *[]string) ForeignUserClaims {
	return ForeignUserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   discordUserID, //* Issued to a discord user
			Issuer:    "gateway",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			Audience:  []string{"gateway"},
		},
		DiscordClaims: DiscordClaims{
			DiscordServerID: discordServerID, //* Make token Discord-server-scoped
			Permissions:     permissions,
		},
		RequestOrigin: DiscordBotRequestOrigin,
	}
}
