package main

import (
	"crypto/rsa"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/gateway/middleware"
	utils "github.com/DomNidy/saint_sim/apps/gateway/utils"
	auth "github.com/DomNidy/saint_sim/pkg/auth"
	tokens "github.com/DomNidy/saint_sim/pkg/auth/tokens"
	saintutils "github.com/DomNidy/saint_sim/pkg/utils"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	gin "github.com/gin-gonic/gin"
)

var log = logging.GetLogger()

var (
	PublicKey  *rsa.PublicKey  = auth.LoadPublicKey("public_key.pem")
	PrivateKey *rsa.PrivateKey = auth.LoadPrivateKey("private_key.pem")
)

func main() {
	userJwt, err := auth.NewForeignUserJWT(PrivateKey, "12345", saintutils.StrPtr("9999"), nil)
	if err != nil {
		log.Printf("Error creating jwt: %v", err)
		return
	}
	log.Printf("User jwt: %v", userJwt)

	validJWT, err := auth.ParseAndIdentifyToken(userJwt, tokens.DiscordBotRequestOrigin, PublicKey)
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
		userClaims, err := utils.GetAuthenticatedData(c.Request.Context())

		if err != nil {
			log.Errorf(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		log.Printf("%v", userClaims)
		switch t := userClaims.(type) {
		case tokens.ForeignUserClaims:
			log.Printf("Foreign user id: %v", t.Subject)
		case tokens.NativeUserClaims:
			log.Printf("Native user id: %v", t.Subject)
		default:
			log.Errorf("Failed to assert claims of ")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}

	})

	r.Run("0.0.0.0:7000")

}
