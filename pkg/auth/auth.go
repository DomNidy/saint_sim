package auth

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Utility function to generate current timestamp in seconds
func CurrentUnixTimestamp() uint32 {
	return uint32(time.Now().Unix())
}

// LoadPrivateKey loads an RSA private key from a PEM file
func LoadPrivateKey(path string) *rsa.PrivateKey {
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read private key file: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		log.Fatalf("failed to parse private key: %v", err)
	}

	return privateKey
}

// LoadPublicKey loads an RSA public key from a PEM file
func LoadPublicKey(path string) *rsa.PublicKey {
	publicKeyBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read public key file: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatalf("failed to parse public key: %v", err)
	}

	return publicKey
}

// SignJWT signs a JWT using RS256 with the provided payload
func SignJWT(payload map[string]interface{}, privateKey *rsa.PrivateKey) (string, error) {
	claims := jwt.MapClaims(payload)

	// Add standard claims if not provided
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = CurrentUnixTimestamp()
	}
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = CurrentUnixTimestamp() + 1*3600 // Default expiration: 1 hour
	}

	// Signing options
	// Create a new token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token using the private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// VerifyJWT verifies a JWT using RS256 with the public key
func VerifyJWT(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return token, nil
}
