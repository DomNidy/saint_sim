package main

import (
	"crypto/rsa"
	"errors"
	"net/http"

	"github.com/DomNidy/saint_sim/apps/gateway/middleware"
	auth "github.com/DomNidy/saint_sim/pkg/auth"
	"github.com/DomNidy/saint_sim/pkg/utils"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	gin "github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var log = logging.GetLogger()

var (
	PublicKey  *rsa.PublicKey  = auth.LoadPublicKey("public_key.pem")
	PrivateKey *rsa.PrivateKey = auth.LoadPrivateKey("private_key.pem")
)

// Extracts the user JWT from request context, and asserts that it satisfies the `JWTPayload` interface
// In order to extract the concrete type, you will need to perform interface type assertion to
// NativeUserJWTClaims and ForeignUserJWTClaims
func getUser(c *gin.Context) (auth.JWTPayload, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, errors.New("reached endpoint but no user is stored in request context")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return nil, errors.New("failed to assert user stored in request context to jwt token type")
	}

	// TODO: Error, the assertion will fail here because token.Claims is just a map containing the claims--
	// TODO: it's not actually a custom struct, so it fails to satisfy the auth.JWTPayload interface.
	// TODO: To fix this, we'd need to try and unmarshal token.Claims to the foreign user claim type and then
	// TODO: native user claim type if the first unmarshal fails.
	// var foreignUserClaims auth.ForeignUserJWTClaims

	userClaims, ok := token.Claims.(auth.JWTPayload)
	if !ok {
		log.Debugf("Failed to assert user claims to JWTPayload interface: %v", token.Claims)
		return nil, errors.New("failed to assert user claims to JWTPayload interface")
	}

	return userClaims, nil

}

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
		userClaims, err := getUser(c)

		if err != nil {
			log.Errorf(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if user, ok := userClaims.(*auth.ForeignUserJWTClaims); ok {
			log.Printf("Found ForeignUserJWTClaims! %s", user.Subject)
		} else if user, ok := userClaims.(*auth.NativeUserJWTClaims); ok {
			log.Printf("Found NativeUserJWTClaims! %s", user.Subject)
		} else {
			log.Errorf("Failed to assert user claims from JWTPayload type to a concrete type.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

	})

	r.Run("0.0.0.0:7000")

}
