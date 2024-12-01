package main

import (
	"crypto/rsa"
	"log"

	"github.com/DomNidy/saint_sim/apps/gateway/middleware"
	auth "github.com/DomNidy/saint_sim/pkg/auth"
	gin "github.com/gin-gonic/gin"
)

var (
	PublicKey  *rsa.PublicKey  = auth.LoadPublicKey("public_key.pem")
	PrivateKey *rsa.PrivateKey = auth.LoadPrivateKey("private_key.pem")
)

func main() {

	jwtClaims := auth.ForeignUserJWTPayload{
		Subject:         "Domaz",
		Issuer:          "gateway",
		IssuedAt:        auth.CurrentUnixTimestamp(),
		Expiration:      auth.CurrentUnixTimestamp() + 4*3600,
		DiscordUserID:   "218526317988151307",
		DiscordServerID: "",
		Permissions:     []string{"a", "b"},
	}

	jwtMappedClaims, err := jwtClaims.ToMap()
	if err != nil {
		log.Printf("Failed to map jwt claims: %v", err)
		return
	}

	signedJwt, err := auth.SignJWT(jwtMappedClaims, PrivateKey)

	if err != nil {
		log.Printf("Error signing jwt: %v", err)
		return
	}

	log.Printf("Signed jwt: %v", signedJwt)

	validJWT, err := auth.VerifyJWT(signedJwt, PublicKey)
	if err != nil {
		log.Printf("Error validating jwt: %v", err)
		return

	}

	log.Printf("Validated jwt: %v", validJWT)
	log.Printf("jwt: %v", signedJwt)

	r := gin.Default()

	authorized := r.Group("/", middleware.Authenticate(PublicKey))

	authorized.GET("/health", func(c *gin.Context) {
		c.JSON(
			200, gin.H{
				"status": "healthy",
			},
		)
	})
	r.Run("0.0.0.0:7000")

}
