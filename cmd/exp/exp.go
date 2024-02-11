package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")
var ErrLiar = errors.New("liar")

func A() error {
	return ErrLiar
}

func B() error {
	err := A()
	if err != nil {
		return fmt.Errorf("b: %w", err)
	}
	return nil
}

func main() {
	err := B()
	if errors.Is(err, ErrNotFound) {
		fmt.Println(err)
	}
}