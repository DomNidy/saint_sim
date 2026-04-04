// Package middleware provides middleware that is used to authenticate incoming API requests
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/DomNidy/saint_sim/apps/api/api_utils"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
)

var (
	errAPIKeySanityCheckFail = errors.New("api key sanity check failed")
	errInvalidAPIKey         = errors.New("invalid api key")
)

// AuthRequire validates that incoming requests provide a valid Api-Key.
func AuthRequire(dbClient dbqueries.Queries) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		apiKey := ginContext.GetHeader("Api-Key")
		if apiKey == "" {
			log.Printf("No API key provided in request")
			ginContext.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})

			return
		}

		err := checkAPIKeyExists(ginContext, dbClient, apiKey)
		if err != nil {
			log.Printf("Error occurred while checking request's API key: %v", err)

			if errors.Is(err, errInvalidAPIKey) {
				ginContext.AbortWithStatusJSON(
					http.StatusForbidden,
					gin.H{"error": "Invalid API key"},
				)

				return
			}

			ginContext.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error"},
			)

			return
		}

		ginContext.Next()
	}
}

// check if a raw (unhashed) API key exists in the db.
// Call with API keys received directly from users; key is hashed internally and looked up.
func checkAPIKeyExists(c context.Context, dbClient dbqueries.Queries, rawAPIKey string) error {
	hashedAPIKey := api_utils.HashApiKey(rawAPIKey)

	resAPIKey, err := dbClient.GetApiKey(c, hashedAPIKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidAPIKey
		}

		return fmt.Errorf("error occurred while looking up API key: %w", err)
	}

	// sanity check
	if resAPIKey.ApiKey != hashedAPIKey {
		return errAPIKeySanityCheckFail
	}

	return nil
}
