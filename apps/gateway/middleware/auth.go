package middleware

import (
	"crypto/rsa"
	"net/http"
	"strings"

	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"

	auth "github.com/DomNidy/saint_sim/pkg/auth"
	gin "github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var log = logging.GetLogger()

// Middleware that authenticates requests from both native and foreign users
func Authenticate(publicKey *rsa.PublicKey) func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Extract the Authorization header prefix and token strings (split them by space)
		headerPrefix, tokenString, found := strings.Cut(authHeader, " ")

		// If the Authorization header wasn't formatted properly (separated by spaces)
		if !found {
			log.Warnf("Failed to parse Authorization header: %v", authHeader)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Make sure token was signed by saint back-end
		token, err := auth.VerifyJWT(tokenString, publicKey)
		if err != nil {
			log.Errorf(err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Errorf("Failed to extract token claims")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Determine the origin of request from the authorization header prefix & token fields
		var requestOrigin auth.RequestOrigin
		if headerPrefix == "Bot" {
			requestOrigin = auth.DiscordBotRequestOrigin
		} else if headerPrefix == "Bearer" {
			requestOrigin = auth.WebRequestOrigin
		} else {
			log.Errorf("Received unrecognized prefix for Authorization header")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Check if the provided JWT is valid for the origin of this request
		// This is more for consistency & organization rather than security
		validForOrigin, err := auth.IsJWTValidForOrigin(claims, requestOrigin)
		if !validForOrigin {
			log.Errorf("Request from origin %v, cannot be authenticated with provided token (token is not valid for this origin)", requestOrigin)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		} else if err != nil {
			log.Errorf("Error while checking if token is valid for origin: %v", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		log.Debugf(token.Raw)
		// Store the token in request context
		c.Set("user", token)
		c.Next()
	}

}
