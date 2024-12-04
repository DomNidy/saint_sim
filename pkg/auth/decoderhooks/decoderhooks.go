package decoderhooks

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	tokens "github.com/DomNidy/saint_sim/pkg/auth/tokens"
	"github.com/DomNidy/saint_sim/pkg/utils"
	logging "github.com/DomNidy/saint_sim/pkg/utils/logging"
	"github.com/golang-jwt/jwt/v4"
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
//		type SomeStructToDecodeInto struct {
//		    Username string `json:"user_name"`,
//		    Age int `json:"user_age"`,
//		}
//
//		if correspondingFieldPairs, err := getJsonPropertyNameStructFieldNameMapping(reflect.TypeOf(SomeStructToMapTo))
//	 if err != nil { ... return }
//	 correspondingFieldPairs will be a map: { "user_name":"Username", "user_age":"Age"}
func getJsonPropertyNameStructFieldNameMapping(structType reflect.Type) (map[string]string, error) {
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("tried to get json property mapping for a type that was not a struct kind")
	}

	fields := reflect.VisibleFields(structType)

	// Maps JSON tags to the name of the struct field
	// they were found in (e.g., "user_name":"Username")
	jsonTagToStructFieldName := make(map[string]string)

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
		log.Debugf("type of data: %v", reflect.TypeOf(data))
		log.Debugf("data: %v", data)

		// Make sure that we're reflecting from Map kind to a
		// struct kind (`MapClaims` to `ForeignUserClaims`)
		if f != reflect.Map || t != reflect.Struct {
			return data, nil
		}

		dataMap, ok := data.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("failed to assert input data to jwt.MapClaims")
		}

		foreignUserClaims := &tokens.ForeignUserClaims{}
		claimStructValue := reflect.ValueOf(foreignUserClaims).Elem()

		jsonPropertyNameToStructFieldName, err := getJsonPropertyNameStructFieldNameMapping(reflect.TypeOf(*foreignUserClaims))

		if err != nil {
			return nil, err
		}

		for jsonPropertyName, fieldName := range jsonPropertyNameToStructFieldName {
			log.Debugf("Mapping: '%v' -> '%v'", jsonPropertyName, fieldName)
			targetFieldValue := claimStructValue.FieldByName(fieldName)

			if !targetFieldValue.IsValid() {
				return nil, fmt.Errorf("failed to find field with name '%v' in ForeignUserClaims struct type", fieldName)
			}

			if !targetFieldValue.CanSet() {
				return nil, fmt.Errorf("cannot set field with name '%v' in ForeignUserClaims struct type", fieldName)
			}

			inputValueForField, exists := dataMap[jsonPropertyName]
			if !exists {
				log.Warnf("`ForeignUserClaims` struct type has a struct field tag that maps field '%v' to json property '%v', so the input data map should have the key `%v`, but it did not", fieldName, jsonPropertyName, jsonPropertyName)
				continue
			}

			// Type of the value we are trying to set
			targetType := targetFieldValue.Type()

			// Perform type conversions to map input values to the types
			// excepted by `ForeignUserClaims`
			switch targetType {
			case reflect.TypeOf((*jwt.NumericDate)(nil)):
				floatVal, ok := inputValueForField.(float64)
				if !ok {
					return nil, fmt.Errorf("expected float64 for json property '%v', but got '%T'", jsonPropertyName, inputValueForField)
				}

				convertedValue := jwt.NewNumericDate(time.Unix(int64(floatVal), 0))
				targetFieldValue.Set(reflect.ValueOf(convertedValue))
				log.Printf("Converted and set %v -> %v", jsonPropertyName, convertedValue)
				continue
			case reflect.TypeOf((*string)(nil)):
				strVal, ok := inputValueForField.(string)
				if !ok {
					return nil, fmt.Errorf("expected string for json property '%v', but got '%T'", jsonPropertyName, inputValueForField)
				}

				convertedValue := utils.StrPtr(strVal)
				targetFieldValue.Set(reflect.ValueOf(convertedValue))
				log.Printf("Converted and set %v -> %v", jsonPropertyName, convertedValue)
				continue

			case reflect.TypeOf(tokens.RequestOrigin("")):
				strVal, ok := inputValueForField.(string)
				if !ok {
					return nil, fmt.Errorf("expected string for json property '%v', but got '%T'", jsonPropertyName, inputValueForField)
				}

				convertedValue := tokens.RequestOrigin(strVal)
				targetFieldValue.Set(reflect.ValueOf(convertedValue))
				log.Printf("Converted and set %v -> %v", jsonPropertyName, convertedValue)
				continue
			case reflect.TypeOf(jwt.ClaimStrings{}):
				sliceVal, ok := inputValueForField.([]interface{})
				if !ok {
					return nil, fmt.Errorf("expected jwt.ClaimStrings for json property '%v', but got '%T'", jsonPropertyName, inputValueForField)
				}

				// Convert []interface{} to []string
				strSlice := make([]string, len(sliceVal))
				for i, v := range sliceVal {
					strVal, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("expected element of jwt.ClaimStrings to be a string, but got '%T'", v)
					}
					strSlice[i] = strVal
				}

				// Assign the converted value
				convertedValue := jwt.ClaimStrings(strSlice)
				targetFieldValue.Set(reflect.ValueOf(convertedValue))
				log.Printf("Converted and set %v -> %v", jsonPropertyName, convertedValue)
				continue
			}

			// Get the reflect.Value of the value to be set
			log.Printf("Target field val: %v", targetFieldValue)
			if targetFieldValue.Type() != reflect.TypeOf(inputValueForField) {
				return nil, fmt.Errorf("the key in the input map '%v' had value of type '%v', but the corresponding field `ForeignClaims.%v` expects a value of type %v", jsonPropertyName, reflect.TypeOf(inputValueForField), fieldName, targetFieldValue.Type())
			}

			// Need to set foreignUserClaims.{struct field name} equal to inputValueForField
			targetFieldValue.Set(reflect.ValueOf(inputValueForField))
			log.Printf("Set %v", fieldName)
		}

		return (interface{})(foreignUserClaims), nil

	}
}
