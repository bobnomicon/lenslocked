package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Password string
}

func DefaultSMTPConfig() SMTPConfig {
	// Get the int value
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		panic(err)
	}

	// SMTP config
	return SMTPConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: port,
		User: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Create email message fields
	from := "test@lenslocked.com"
	to := "miller.robert.john@gmail.com"
	subject := "This is a test email"
	plaintext := "This is the body of the email."
	html := `<h1>This is a header</h1><p>This is a paragraph.</p><p>This is another paragraph.</p>`
	msg := gomail.NewMessage()
	msg.SetHeader("To", to)
	msg.SetHeader("From", from)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", plaintext)
	msg.AddAlternative("text/html", html)

	// Send the email
	cfg := DefaultSMTPConfig()
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	if err := dialer.DialAndSend(msg); err != nil {
		panic(err)
	}
}
