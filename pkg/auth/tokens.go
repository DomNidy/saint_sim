package auth

import (
	"encoding/json"
	"fmt"
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

// JWT that native users use to authenticate.
// This is presented in a HTTP header as: "Authorization: Bearer <jwt>"
//
// This is distinct from the `ForeignUserJWTPayload`
// used by the Saint Discord bot.
//
// A native user refers to an end user that directly
// interacts with the Saint API (i.e. their requests aren't
// proxied through the Saint Discord bot). This is in contrast
// to a foreign user.
type NativeUserJWTPayload struct {
	// Native user ID
	Subject string `json:"sub"` // The unique identifier for the user

	// Issuer of the token (e.g., gateway)
	Issuer string `json:"iss"`

	// IssuedAt is the timestamp when the token was issued
	IssuedAt uint32 `json:"iat"`

	// Expiration is the timestamp when the token expires
	Expiration uint32 `json:"exp"`

	// Used to identify the origin (Discord or web app) a token
	// is valid for.
	RequestOrigin RequestOrigin `json:"request_origin"`
}

// JWT that foreign users will be authenticated with.
// This is presented in a HTTP header as: "Authorization: Bot <jwt>"
//
// A foreign user refers to an end user that
// interacts with the Saint API through the
// Saint Discord bot. This is distinct from the
// `NativeUserJWTPayload` used by native users.
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
type ForeignUserJWTPayload struct {
	// Discord user ID
	Subject string `json:"sub"` // The unique identifier for the Discord user

	// Issuer of the token (e.g., gateway)
	Issuer string `json:"iss"`

	// IssuedAt is the timestamp when the token was issued
	IssuedAt uint32 `json:"iat"`

	// Expiration is the timestamp when the token expires
	Expiration uint32 `json:"exp"`

	// Used to identify the origin (Discord or web app) a token
	// is valid for.
	RequestOrigin RequestOrigin `json:"request_origin"`

	// Custom claims for Discord-specific context
	DiscordServerID string `json:"discord_server_id,omitempty"` // The server ID for scoping access

	DiscordUserID string `json:"discord_user_id,omitempty"` // Optional specific user ID for scoping access

	Permissions []string `json:"permissions,omitempty"` // Optional permissions granted to this token
}

type JWTPayload interface {
	// Return the JWT payload as map
	ToMap() (map[string]interface{}, error)
}

// This method converts a `ForeignUserJWTPayload` type to map[string]interface{}
//
// This is needed as the golang-jwt library expects the payload to be of type map[string]interface{}
// To perform this conversion, we simply marshal the struct into json, then unmarshal it
// into map[string]interface{} type. This method uses the struct tags of `ForeignUserJWTPayload`
// to perform marshalling and unmarshalling.
func (t *ForeignUserJWTPayload) ToMap() (map[string]interface{}, error) {
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

// This method converts a `NativeUserJWTPayload` type to map[string]interface{}
func (t *NativeUserJWTPayload) ToMap() (map[string]interface{}, error) {
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