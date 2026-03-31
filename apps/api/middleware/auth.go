package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
)

// Middleware: Authenticates requests
func AuthRequire(db *dbqueries.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Api-Key")
		if apiKey == "" {
			log.Printf("No API key provided in request")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		hashedApiKey := api_utils.HashApiKey(apiKey)
		serviceName, err := db.GetApiKeyServiceName(c, hashedApiKey)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Printf(
					"Auth middleware -- invalid API key provided. Cannot authenticate request.",
				)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
				return
			}
			log.Printf("Error occured in auth middleware while trying to validate API key: %s", err)
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error occurred"},
			)
			return
		}

		// Ensure the api key is authorized for this service
		if serviceName != "api" {
			log.Printf(
				"api key '%s' was issued for a service other than 'api': '%s'",
				apiKey,
				serviceName,
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
