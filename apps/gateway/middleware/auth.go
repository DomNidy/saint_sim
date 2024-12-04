package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"

	auth "github.com/DomNidy/saint_sim/pkg/auth"
	tokens "github.com/DomNidy/saint_sim/pkg/auth/tokens"
	gin "github.com/gin-gonic/gin"
)

var log = logging.GetLogger()

// Authenticated user/service data will be stored in the gin context using this key (after Authenticate middleware is run)
type authenticatedDataKey string

const AuthedDataKey authenticatedDataKey = "authenticated_data"

// Middleware that authenticates requests from both native and foreign users
func Authenticate(publicKey *rsa.PublicKey) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Figure out what origin created this requet (web app, discord bot, etc.)
		requestOrigin, err := getRequestOrigin(c)
		if err != nil {
			log.Debug(err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Extract the Authorization header prefix and token strings (split them by space)
		authHeader := c.GetHeader("Authorization")
		_, tokenString, found := strings.Cut(authHeader, " ")

		// If the Authorization header wasn't formatted properly (separated by spaces)
		if !found {
			log.Warnf("Failed to parse Authorization header: %v", authHeader)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Make sure token is valid (was signed by Saint back-end)
		claims, err := auth.ParseAndIdentifyToken(tokenString, requestOrigin, publicKey)
		if err != nil {
			log.Errorf(err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Store the token in request context
		ctx := context.WithValue(c.Request.Context(), AuthedDataKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}

}

// Look at the request and determine its request origin based on the
// Authorization header.
func getRequestOrigin(c *gin.Context) (tokens.RequestOrigin, error) {
	authHeader := c.GetHeader("Authorization")

	prefix := strings.Split(authHeader, " ")

	if len(prefix) == 0 {
		return "", fmt.Errorf("failed to extract request origin field from token")
	}

	// For now, we'll just assume the prefix "Bot" indicates a discord bot request
	// This means that any service which uses the "Bot" prefix
	if prefix[0] == "Bot" {
		return tokens.DiscordBotRequestOrigin, nil
	} else if prefix[0] == "Bearer" {
		return tokens.WebRequestOrigin, nil
	}

	return "", fmt.Errorf("unrecognized prefix (authentication scheme). expected 'Bearer' or 'Bot")
}
