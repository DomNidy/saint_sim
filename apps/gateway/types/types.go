package types

import "net/http"

// Main object that contains multiple gateway routes.
// Does not explicitly NEED a request origin adapater to function
type GatewayRouter struct {
}

// A Gateway
type GatewayRoute struct {
}

// Wrapper around gin api request matching
type MatchGatewayRequest struct {
}

// String that identifies the front-end a request originated from.
//
// A `RequestOrigin` can be extracted from an HTTP request. It is an identifier
// that details which front-end a request originated from (i.e., "web", or "discord_bot").
type GatewayRequestOrigin string

// Adds a route to the gateway. Each route should be able to receive
// requests from arbitrary many `RequestOrigin`s, and handle them differently.
// func (r *GatewayRouter) AddGatewayRoute()

// Receive HTTP request and parse it's 'SaintRequestOrigin' header
// to determine the request origin (e.g., "discord_bot", "web", etc.)
type GatewayOriginAdapter func(request http.Request) GatewayRequestOrigin
