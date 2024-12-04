package decoderhooks

import (
	"fmt"
	"reflect"
	"strings"

	tokens "github.com/DomNidy/saint_sim/pkg/auth/tokens"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	"github.com/mitchellh/mapstructure"
)

var log = logging.GetLogger()

// Verifies that each key of `mapKeys` uniquely corresponds to a struct field JSON
// tag in the `structType` struct type. If the mapping satisfies bijectivity,
// a map is returned where each key:value pair is a jsonTag:StructFieldName.
// The key 'jsonTag' is a struct field JSON tag which was matched with a key of the `mapKeys`.
// The value 'StructFieldName' is the name of the field of the `structType` struct. (which is
// tagged with 'jsonTag').
//
// The relation between keys of `mapKeys` and the tags of `structType` should be bijective
// (one-to-one correspondance).
//
// This should be called in decoder hook functions and passed a list of the keys
// in the decoder's input data map, as well as the struct type the decoder hook decodes to.
//
// Note: The uniqueness property of bijectivity is upheld by the Go compiler (Meaning you
// will receive warnings at compile time if a struct maps multiple fields to the same json tag)
//
// Example:
// ...in decoder hook fn
//
// - The data we are mapping from
//
//	var someDataMap := map[string]interface{
//	     "user_name": "John",
//			"user_age":12,
//	}
//
// - The type we are decoding to
//
//	type SomeStructToDecodeInto struct {
//	    Username string `json:"user_name"`,
//	    Age int `json:"user_age"`,
//	}
//
//	if correspondingFieldPairs, err := checkMapKeysAndStructJsonTagsAreBijective([]string{"user_name"}, reflect.TypeOf(SomeStructToMapTo)); err != nil {
//	    ...in this example, this will execute because `mapKeys` forgot to include the `"user_age"` key,
//	    ...which is expected because of the `Age` field's JSON tag in `SomeStructToMapTo`
//	}
func checkMapKeysAndStructJsonTagsAreBijective(mapKeys []string, structType reflect.Type) (map[string]string, error) {
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("bijection check was called on a type which was not a struct kind")
	}

	fields := reflect.VisibleFields(structType)

	// Maps JSON tags to the name of the struct field
	// they were found in (e.g., "user_name":"Username")
	jsonTagToStructFieldName := make(map[string]string, len(mapKeys))

	// For each field in structType, read it's tags
	for _, field := range fields {
		fieldJsonTag := field.Tag.Get("json")

		if fieldJsonTag == "" {
			log.Debugf("'%v' field has no json tag", field.Name)
			continue
		}

		// parse out the tag name by removing 'omitempty' if it exists
		parsedJsonTag := strings.TrimSuffix(fieldJsonTag, ",omitempty")

		// parse out the tag name
		log.Debugf("'%v' field has json tag: %v", field.Name, parsedJsonTag)

		jsonTagToStructFieldName[parsedJsonTag] = field.Name
	}

	return jsonTagToStructFieldName, nil
}

// Helper function to format type mismatches in decoder hooks
// var typeConvErr = func(key string, val interface{}, expectedType reflect.Kind) error {
// 	return fmt.Errorf("'%v' field in input data failed to be asserted to '%v' type. The field's value was: %v", key, expectedType, val)
// }

// Custom mapstructure decoder hook used to convert
// `jwt.MapClaims` struct to the `ForeignUserClaims` struct type.
func DecodeJwtMapClaimsToForeignUserClaims() mapstructure.DecodeHookFuncKind {
	return func(f, t reflect.Kind, data interface{}) (interface{}, error) {
		log.Debugf("f kind: %v", f)
		log.Debugf("t type: %v", reflect.TypeOf(t))
		log.Debugf("data: %v", data)
		log.Debugf("type of data: %v", reflect.TypeOf(data))

		// Make sure that we're reflecting from Map kind to a
		// struct kind (`MapClaims` to `ForeignUserClaims`)
		if f != reflect.Map || t != reflect.Struct {
			return data, nil
		}

		var foreignUserClaims tokens.ForeignUserClaims

		correspondingFieldPairs, err := checkMapKeysAndStructJsonTagsAreBijective([]string{"discord_server_id"}, reflect.TypeOf(foreignUserClaims))

		if err != nil {
			return nil, err
		}

		for jsonTagName, fieldName := range correspondingFieldPairs {
			log.Debugf("Mapping: '%v' -> '%v'", jsonTagName, fieldName)
		}

		// dataMap, ok := data.(jwt.MapClaims)
		// if !ok {
		// 	return nil, errors.New("failed to assert input data to `jwt.MapClaims` type")
		// }
		// for key, val := range dataMap {
		// 	log.Debugf("%v : %v", key, val)
		// 	switch key {
		// 	case "discord_server_id":
		// 		v, ok := val.(string)

		// 		if !ok {
		// 			return nil, typeConvErr(key, val, reflect.String)
		// 		}

		// 		foreignUserClaims.DiscordServerID = &v

		// 	case "request_origin":
		// 		v, ok := val.(string)

		// 		if !ok {
		// 			return nil, typeConvErr(key, val, reflect.String)
		// 		}

		// 		foreignUserClaims.RequestOrigin = tokens.RequestOrigin(v)
		// 	default:
		// 		return nil, fmt.Errorf("failed to match a case for the key '%v' of the pair '%v:%v' in the input data. Make sure the input map's field names match the struct field tags for the `ForeignUserClaims` type", key, key, val)
		// 	}
		// }

		return (interface{})(foreignUserClaims), nil

	}
}
