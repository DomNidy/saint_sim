package decoderhooks

import (
	"fmt"
	"reflect"
	"time"

	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mitchellh/mapstructure"
)

var log = logging.GetLogger()

var foreignUserClaimsNameMapping = map[string]string{
	"discord_server_id": "DiscordServerID",
	"permissions":       "Permissions",
	"request_origin":    "RequestOrigin",
	"iss":               "Issuer",
	"sub":               "Subject",
	"aud":               "Audience",
	"exp":               "ExpiresAt",
	"nbf":               "NotBefore",
	"iat":               "IssuedAt",
	"jti":               "ID",
}

// Function used to match map keys to struct field names for
// the `ForeignUserClaims` struct
func ForeignUserClaimsMatchNameFn(mapKey, fieldName string) bool {
	expectedFieldName, exists := foreignUserClaimsNameMapping[mapKey]

	if !exists {
		log.Errorf("Received a mapKey that was not present in name mapping '%v'", mapKey)
	}

	return fieldName == expectedFieldName
}

// Helper function to format type mismatches in decoder hooks
// var typeConvErr = func(key string, val interface{}, expectedType reflect.Kind) error {
// 	return fmt.Errorf("'%v' field in input data failed to be asserted to '%v' type. The field's value was: %v", key, expectedType, val)
// }

// Custom mapstructure decoder hook used to convert
// `jwt.MapClaims` struct to the `ForeignUserClaims` struct type.
func DecodeJwtMapClaimsToForeignUserClaims() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f == reflect.TypeOf(float64(0)) && t == reflect.TypeOf(jwt.NumericDate{}) {
			dataFloat, ok := data.(float64)
			if !ok {
				return data, fmt.Errorf("failed to assert data (%v) to float64 type, so we can't convert this value to jwt.NumericDate", data)
			}
			return jwt.NewNumericDate(time.Unix(int64(dataFloat), 0)), nil
		}

		return data, nil
	}
}
