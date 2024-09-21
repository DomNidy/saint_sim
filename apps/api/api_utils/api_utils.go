package api_utils

import (
	uuid "github.com/google/uuid"
)

// Use to generate UUID for simulation operations & results
func GenerateUUID() string {
	return uuid.New().String()
}
