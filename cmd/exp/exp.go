package main

import (
	"errors"
	"fmt"
	"log"
)

func Connect() error {
	// Try to connect
	// Pretend we got an error
	defer fmt.Println("Deferred!")
	err := errors.New("connection failed")
	fmt.Println(err)
	return err
}

func CreateUser() error {
	err := Connect()
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func CreateOrg() error {
	err := CreateUser()
	if err != nil {
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

func main() {
	err := CreateUser()
	if err != nil {
		log.Println(err)
	}

	err = CreateOrg()
	if err != nil {
		log.Println(err)
	}
}