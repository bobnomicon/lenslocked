package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type ctxKey string

const (
	favoriteColorKey ctxKey = "favorite-color"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, favoriteColorKey, "blue")
	anyValue := ctx.Value(favoriteColorKey) // Return type 'any'

	stringValue, ok := anyValue.(string)
	if !ok {
		fmt.Println(anyValue, "is not a string!")
		return
	}

	fmt.Println(strings.HasPrefix(stringValue, "b"))

	// Type conversion/assertion
	fmt.Printf("Type of '%v': %T\n", anyValue, anyValue) // '%T' is the type
	fmt.Println(reflect.TypeOf(anyValue)) // Returns the dynamic type

	var a any = "hello"
	s, ok := a.(string) // 'ok' is an optional boolean, is true if type assertion successful
	fmt.Println(s, ok) // s' has the value of the new type (string) if successful
	
	i := a.(int) // If omit 'ok' and type assertion fails, will panic
	fmt.Println(i)
}
