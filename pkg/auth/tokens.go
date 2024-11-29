package auth

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
// - discord-server-scoped JWTs (access control based
//      on the originating discord server)
// - discord-user-scoped JWTs (access control based
//      on specific discord user id)
type ForeignUserJWTPayload struct {
	// Discord user ID or bot ID
	Subject string `json:"sub"` // The unique identifier for the Discord user/bot

	// Issuer of the token (e.g., Saint API)
	Issuer string `json:"iss"`

	// IssuedAt is the timestamp when the token was issued
	IssuedAt uint32 `json:"iat"`

	// Expiration is the timestamp when the token expires
	Expiration uint32 `json:"exp"`

	// Custom claims for Discord-specific context
	DiscordServerID string `json:"discord_server_id"` // The server ID for scoping access

	DiscordUserID string `json:"discord_user_id,omitempty"` // Optional specific user ID for scoping access

	Permissions []string `json:"permissions,omitempty"` // Optional permissions granted to this token
}
