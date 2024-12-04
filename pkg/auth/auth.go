package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/DomNidy/saint_sim/pkg/auth/decoderhooks"
	"github.com/DomNidy/saint_sim/pkg/auth/tokens"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mitchellh/mapstructure"
)

var log = logging.GetLogger()

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

// Creates and signs a JWT using `privateKey`, the string of the token is returned.
// This key is valid for requests initiated by foreign users (Discord users).
// If an error occurs, the returned string is empty ("") and the error will be non nil.
func NewForeignUserJWT(privateKey *rsa.PrivateKey, discordUserID string, discordServerID *string, permissions *[]string) (string, error) {
	claims := tokens.CreateDiscordUserClaims(discordUserID, discordServerID, permissions)
	mappedClaims, err := claims.ToMap()

	if err != nil {
		return "", fmt.Errorf("error mapping claims: %v", err)
	}

	signedJwt, err := SignJWT(mappedClaims, privateKey)
	if err != nil {
		return "", fmt.Errorf("error signing jwt: %v", err)
	}

	return signedJwt, nil
}

// `ParseAndIdentifyToken` parses and verifies a JWT token string,  indentifies what type
// of token it is (e.g., `NativeUserClaims` or `ForeignUserClaims`), and then maps it to
// the corresponding token type struct.
//
// This function also verifies that the passed token can be used to authenticate
// requests from the indicated request origin. (e.g., `ForeignUserClaims` tokens should
// only be valid for requests that originate from the `DiscordBotRequestOrigin` request origin)
//
// The returned value is typed as `interface{}`, as the caller should perform type
// assertion to get the concrete type & value.
func ParseAndIdentifyToken(tokenString string, requestOrigin tokens.RequestOrigin, publicKey *rsa.PublicKey) (interface{}, error) {
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

	// We need to first perform interface assertion
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to assert token.Claims to jwt.MapClaims type")
	}
	log.Printf("ParseAndIdentifyToken extracted claims: %v", claims)

	// First, differentiate between the two claim types by reading their `request_origin` fields
	tokenOrigin, err := getTokenOrigin(claims)
	if err != nil {
		log.Warnf("Received a token which was successfully validated, but did not have a request_origin field. This implies that we issued this token ourselves, and forgot to include the request_origin field. Review the token generation functionality.")
		return nil, errors.New(err.Error())
	}

	if requestOrigin != tokenOrigin {
		log.Errorf("A token was received from request origin '%v', but the token's `request_origin` field indicates the token is only valid for request origin '%v'", requestOrigin, tokenOrigin)
		return nil, fmt.Errorf("token's origin '%v' does not match the provided request origin '%v'", tokenOrigin, requestOrigin)
	}

	log.Debugf("Found token origin: %v", tokenOrigin)
	log.Debugf("Token's claims: %v", claims)

	// Convert the token's claims to the correct struct type based on the `request_origin` claim
	switch tokenOrigin {
	case tokens.DiscordBotRequestOrigin:
		var foreignUserClaims tokens.ForeignUserClaims

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     &foreignUserClaims,
			DecodeHook: decoderhooks.DecodeJwtMapClaimsToForeignUserClaims(),
			MatchName:  decoderhooks.ForeignUserClaimsMatchNameFn,
		})

		if err != nil {
			log.Fatalf("error initializing mapstructure decoder: %v", err)
		}

		if err := decoder.Decode(claims); err != nil {
			log.Errorf(err.Error())
		}

		log.Debugf("Decoded token into ForeignUserClaims struct, data: %v", foreignUserClaims)

		return foreignUserClaims, nil
	case tokens.WebRequestOrigin:
		return nil, fmt.Errorf("decoder not implemented for this token type")
	default:
		return nil, fmt.Errorf("token had unrecognized `request_origin` '%v'", tokenOrigin)
	}

}

// Returns the `request_origin` field from the token as a string, error if it wasn't present
func getTokenOrigin(claims jwt.MapClaims) (tokens.RequestOrigin, error) {
	tokenOrigin, ok := claims["request_origin"]
	if !ok {
		return "", fmt.Errorf("token's claims had no request_origin field")
	}

	origin, ok := tokenOrigin.(string)
	if !ok {
		return "", fmt.Errorf("token's `request_origin` field was present, but it could not be asserted to string type")
	}

	return tokens.RequestOrigin(origin), nil
}
