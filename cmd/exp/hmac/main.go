package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	secretKey := "secret-key"
	cookie := `{"id": 123, "email": "robert@robertmiller.com"}`

	// Setup hashing function using HMAC
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write data to hashing function
	h.Write([]byte(cookie))

	// Get the resulting hash
	result := h.Sum(nil)

	// Resulting hash is binary, so need to hex encode it
	fmt.Println(hex.EncodeToString(result))
}
