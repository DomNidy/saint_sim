// Package interfaces provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package interfaces

const (
	SaintBotAuthScopes = "SaintBotAuth.Scopes"
	UserAuthScopes     = "UserAuth.Scopes"
)

// Defines values for WowCharacterRealm.
const (
	Draenor    WowCharacterRealm = "draenor"
	Hydraxis   WowCharacterRealm = "hydraxis"
	Silvermoon WowCharacterRealm = "silvermoon"
	Thrall     WowCharacterRealm = "thrall"
)

// Defines values for WowCharacterRegion.
const (
	Cn WowCharacterRegion = "cn"
	Eu WowCharacterRegion = "eu"
	Kr WowCharacterRegion = "kr"
	Tw WowCharacterRegion = "tw"
	Us WowCharacterRegion = "us"
)

// ErrorResponse Error response returned by API when something goes wrong
type ErrorResponse struct {
	// Message Message explaining the error
	Message *string `json:"message,omitempty"`
}

// SimulationData The output of a simulation.
type SimulationData struct {
	// FromRequest The ID of the simulation request that initated this simulation
	FromRequest *string `json:"from_request,omitempty"`

	// Id ID of this simulation
	Id *int `json:"id,omitempty"`

	// SimResult The actual data produced from the simulation operation
	SimResult *string `json:"sim_result,omitempty"`
}

// SimulationMessageBody This JSON object is included in a rabbitmq message, then that message gets published to the simulation_queue. Consumers of the simulation queue (simulation_worker) will use this JSON object to carry out the simulation.
type SimulationMessageBody struct {
	// SimulationId Used to identify a simulation request in postgres
	SimulationId *string `json:"simulation_id,omitempty"`
}

// SimulationOptions Specifices sim options, and the character of interest to sim, send this to the api
type SimulationOptions struct {
	// WowCharacter Object containing all data needed to identify a WoW character, used to retrieve their gear and talents, etc. (Realm list here https://worldofwarcraft.blizzard.com/en-us/game/status/us)
	WowCharacter WowCharacter `json:"wow_character"`
}

// SimulationResponse Object containing information about a simulation operation, returned from api
type SimulationResponse struct {
	// SimulationRequestId Used to identify a simulation request in postgres
	SimulationRequestId *string `json:"simulation_request_id,omitempty"`
}

// WowCharacter Object containing all data needed to identify a WoW character, used to retrieve their gear and talents, etc. (Realm list here https://worldofwarcraft.blizzard.com/en-us/game/status/us)
type WowCharacter struct {
	// CharacterName The name of the WoW character
	CharacterName string `json:"character_name"`

	// Realm The realm which the character is located on
	Realm WowCharacterRealm `json:"realm"`

	// Region Identifies the region in which the characters realm is located
	Region WowCharacterRegion `json:"region"`
}

// WowCharacterRealm The realm which the character is located on
type WowCharacterRealm string

// WowCharacterRegion Identifies the region in which the characters realm is located
type WowCharacterRegion string

// InternalError Error response returned by API when something goes wrong
type InternalError = ErrorResponse

// NotFoundError Error response returned by API when something goes wrong
type NotFoundError = ErrorResponse

// SimulateParams defines parameters for Simulate.
type SimulateParams struct {
	// DiscordUserId This header MUST be present for requests originating from the Saint Discord bot. Not required when requested directly from a user.
	DiscordUserId *string `json:"Discord-User-Id,omitempty"`
}

// SimulateJSONRequestBody defines body for Simulate for application/json ContentType.
type SimulateJSONRequestBody = SimulationOptions
