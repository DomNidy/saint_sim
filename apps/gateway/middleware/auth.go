package middleware

import (
	"crypto/rsa"
	"log"
	"net/http"
	"strings"

	types "github.com/DomNidy/saint_sim/apps/gateway/types"
	auth "github.com/DomNidy/saint_sim/pkg/auth"
	gin "github.com/gin-gonic/gin"
)

func Authenticate(publicKey *rsa.PublicKey) func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			log.Printf("Incoming request is missing Authorization header.")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		log.Printf("Received auth header: %v", authHeader)

		// Extract the Authorization header prefix and token strings (split them by space)
		headerPrefix, tokenString, found := strings.Cut(authHeader, " ")

		// If the Authorization header wasn't formatted properly (separated by spaces)
		if !found {
			log.Printf("Failed to parse Authorization header: %v", authHeader)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		// Determine the origin of request from the authorization header prefix
		var requestOrigin types.GatewayRequestOrigin
		if headerPrefix == "Bot" {
			requestOrigin = "discord"
		} else if headerPrefix == "Bearer" {
			requestOrigin = "web"
		} else {
			log.Printf("Received unrecognized prefix for Authorization header")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		token, err := auth.VerifyJWT(tokenString, publicKey)

		if err != nil {
			log.Printf("Failed to validate token with origin %v: %v ", requestOrigin, err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		// Store the token in request context
		c.Set("user", token)
		c.Next()
	}

}
