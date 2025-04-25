package main

import (
	"fmt"

	"github.com/bobnomicon/lenslocked/email"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Email service
	cfg := email.DefaultSMTPConfig()
	es := email.NewEmailService(cfg)
	
	// Send forgot password email
	if err := es.ForgotPassword(
		"miller.robert.john@gmail.com",
		"https://lenslocked/reset-pw?token=abc123",
		nil,
	); err != nil {
		panic(err)
	}
	fmt.Println("Email sent.")
}
