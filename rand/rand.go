package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Returns a byte slice of random bytes using crypto/rand.
// n is the number of bytes to use to generate the random bytes.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	return b, nil
}

// Returns a random string using crypto/rand.
// n is the number of bytes to use to generate the random string.
func String(n int) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Returns a fixed-size (32 bytes) session token.
const SessionTokenBytes = 32
func SessionToken() (string, error) {
	return String(SessionTokenBytes)
}
