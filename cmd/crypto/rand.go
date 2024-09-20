package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	fmt.Println(base64.URLEncoding.EncodeToString(b))
}
