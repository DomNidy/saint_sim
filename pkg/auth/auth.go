package auth

// JWT that native users use to authenticate.
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
	Subject string `json:"sub"`

	IssuedAt uint32 `json:"iat"`
	// TODO: Continue creating type
}

// JWT that foreign users will be authenticated with.
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
	// TODO: Continue creating type
}
