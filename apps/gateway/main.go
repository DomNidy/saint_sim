package main

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
func (r *GatewayRouter) AddGatewayRoute()

// Receive HTTP request and parse it's 'SaintRequestOrigin' header
// to determine the request origin (e.g., "discord_bot", "web", etc.)
type GatewayOriginAdapter func(request http.Request) GatewayRequestOrigin

func main() {
	// 1. Authenticate & authorize (HTTP request -> Authenticated User Token + Request Data)
	//      - this step needs to handle request authentication for arbitrary many `RequestOrigin`s
	// 2. Validate

	// Receive HTTP request and parse it's 'SaintRequestOrigin' header
	// to determine the request origin (e.g., "discord_bot", "web", etc.)
	// Each individual gateway route should be able to override this if it's set at gateway router level
	// gatewayOriginAdapter := func (request http.Request) {
	// 	       switch(request.header["SaintRequestOrigin"])
	//		       case "discord_bot":
	//                 return "discord_bot"
	//		       case "web":
	//                 return "web"
	//		       default:
	//                 return "web"
	//     }

	// Main object that contains multiple gateway routes.
	// gatewayRouter := NewGatewayRouter(gatewayOriginAdapter)
	//
	//
	//
	//
	// gatewayRouter.AddGatewayRoute(
	// 		"GET", 										//* Request method to match
	// 		"/simulate", 								//* Request route to match
	//      { 											//* Auth handlers are different depending on request origin
	//			"discord_bot": AuthDiscordBotRequests,
	// 			"web": AuthWebRequests,
	//      },											//* Actual request handler that performs validation, rate-limiting, caching, and routes the request
	//      func (c *gin.Context, requestOrigin RequestOrigin, user UserToken) {
	//			...
	//		}
	// )
	//
	//
	//
	//
	// r.GET("/simulate", func(c *gin.Context) {
	//
	// })
	//
}
