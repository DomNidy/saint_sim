// Package interfaces provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package interfaces

// Add defines model for Add.
type Add struct {
	A *int64 `json:"a,omitempty"`
	B *int64 `json:"b,omitempty"`
}

// AddResponse defines model for AddResponse.
type AddResponse struct {
	Result *int64 `json:"result,omitempty"`
}

// SimulationOptions defines model for SimulationOptions.
type SimulationOptions struct {
	// CharacterRealm WoW realm of the character we want to sim
	CharacterRealm *string `json:"character_realm,omitempty"`

	// SimConfig Extra configuration options for sims
	SimConfig *struct {
		// TargetCount The amount of enemy targets to include in the sim
		TargetCount *int `json:"target_count,omitempty"`
	} `json:"sim_config"`
}

// SimulationResponse Object containing information about a simulation operation
type SimulationResponse struct {
	SimulationId *string `json:"simulation_id,omitempty"`
}

// AddJSONRequestBody defines body for Add for application/json ContentType.
type AddJSONRequestBody = Add

// SimulateJSONRequestBody defines body for Simulate for application/json ContentType.
type SimulateJSONRequestBody = SimulationOptions