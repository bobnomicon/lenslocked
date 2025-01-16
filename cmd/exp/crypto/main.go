package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mathRand "math/rand"
	"time"
)

func main() {
	// math/rand (use for simulation only i.e. tests)
	// Same seed, same output so use current time as seed
	newMathRand := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	for range [5]int{} {
		fmt.Println(newMathRand.Intn(100))
	}

	// vs. crypto/rand (secure)
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.URLEncoding.EncodeToString(b))
}
