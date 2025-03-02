package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/operas-logicas/lenslocked/models"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Email service
	cfg := models.DefaultSMTPConfig()
	es := models.NewEmailService(cfg)
	
	// Send forgot password email
	if err := es.ForgotPassword(
		"miller.robert.john@gmail.com",
		"https://lenslocked/reset-pw?token=abc123",
	); err != nil {
		panic(err)
	}
	fmt.Println("Email sent.")
}
