package utils

import (
	"errors"
)

var errUnexpectedUUIDValueType = errors.New("uuid was not encoded as string")

// IntPtr returns a pointer to i.
func IntPtr(i int) *int {
	return &i
}

// StrPtr returns a pointer to s.
func StrPtr(s string) *string {
	return &s
}
