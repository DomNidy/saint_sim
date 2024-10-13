package handlers

import (
	"log"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Middleware: Authenticates requests
func AuthRequire(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Api-Key")
		if apiKey == "" {
			log.Printf("No API key provided in request")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		hashedApiKey := api_utils.HashApiKey(apiKey)
		var serviceName string // the service this api key is allowed to auth with (should be 'api')
		err := db.QueryRow(c, "SELECT service_name FROM api_keys WHERE api_key = $1", hashedApiKey).Scan(&serviceName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Ensure the api key is authorized for this service
		if serviceName != "api" {
			log.Printf("api key '%s' was issued for a service other than 'api': '%s'", apiKey, serviceName)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
