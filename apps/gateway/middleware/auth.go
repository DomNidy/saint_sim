package middleware

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"strings"

	auth "github.com/DomNidy/saint_sim/pkg/auth"
	gin "github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Middleware that authenticates requests from both native and foreign users
func Authenticate(publicKey *rsa.PublicKey) func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Extract the Authorization header prefix and token strings (split them by space)
		headerPrefix, tokenString, found := strings.Cut(authHeader, " ")

		// If the Authorization header wasn't formatted properly (separated by spaces)
		if !found {
			log.Printf("Failed to parse Authorization header: %v", authHeader)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		token, err := auth.VerifyJWT(tokenString, publicKey)
		if err != nil {
			log.Printf("Failed to validate token: %v ", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Printf("Failed to extract token claims")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		// Determine the origin of request from the authorization header prefix & token fields
		var requestOrigin auth.RequestOrigin
		if headerPrefix == "Bot" {
			requestOrigin = auth.DiscordBotRequestOrigin
		} else if headerPrefix == "Bearer" {
			requestOrigin = auth.WebRequestOrigin
		} else {
			log.Printf("Received unrecognized prefix for Authorization header")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		// Check if the provided JWT is valid for the origin of this request
		// This is more for consistency & organization rather than security
		validForOrigin, err := isJWTValidForOrigin(claims, requestOrigin)
		if !validForOrigin {
			log.Printf("Request from origin %v, cannot be authenticated with provided token (token is not valid for this origin)", requestOrigin)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		} else if err != nil {
			log.Printf("Error while checking if token is valid for origin: %v", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		}

		// Store the token in request context
		c.Set("user", token)
		c.Next()
	}

}

// Check if a JWT is valid for a request origin
func isJWTValidForOrigin(tokenClaims jwt.MapClaims, origin auth.RequestOrigin) (bool, error) {
	if tokenRequestOrigin, ok := tokenClaims["request_origin"].(string); ok {
		if tokenRequestOrigin == string(origin) {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, fmt.Errorf("Failed to extract request origin field from token")
}