package main

import (
	"crypto/rsa"
	"log"

	"github.com/DomNidy/saint_sim/apps/gateway/middleware"
	auth "github.com/DomNidy/saint_sim/pkg/auth"
	"github.com/DomNidy/saint_sim/pkg/utils"
	gin "github.com/gin-gonic/gin"
)

var (
	PublicKey  *rsa.PublicKey  = auth.LoadPublicKey("public_key.pem")
	PrivateKey *rsa.PrivateKey = auth.LoadPrivateKey("private_key.pem")
)

func main() {
	userJwt, err := auth.NewForeignUserJWT(PrivateKey, "12345", utils.StrPtr("9999"), nil)
	if err != nil {
		log.Printf("Error creating jwt: %v", err)
		return
	}
	log.Printf("User jwt: %v", userJwt)

	validJWT, err := auth.VerifyJWT(userJwt, PublicKey)
	if err != nil {
		log.Printf("Error validating jwt: %v", err)
		return

	}
	log.Printf("Validated jwt: %v", validJWT)

	r := gin.Default()

	authorized := r.Group("/", middleware.Authenticate(PublicKey))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(
			200, gin.H{
				"status": "healthy",
			},
		)
	})

	// Authorized endpoints
	authorized.POST("/simulate", func(c *gin.Context) {
		// user, exists := c.Get("user")
		// TODO: Forward request to `api`
	})

	r.Run("0.0.0.0:7000")

}
