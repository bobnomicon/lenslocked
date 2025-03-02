package models

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

const (
	// Default email to send emails from
	DefaultSender = "support@lenslocked.com"
)

type SMTPConfig struct {
	Host string
	Port int
	Username string
	Password string
}

type Email struct {
	From string
	To string
	Subject string
	Plaintext string
	HTML string
}

type EmailService struct {
	DefaultSender string // Used as the default sender when one isn't provided
	dialer *gomail.Dialer // Unexported field
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
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: gomail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}
	return &es
}

// Used to set the sender for the email message as follows:
//	1. email.From
//	2. EmailService.DefaultSender
//	3. DefaultSender (package const)
func (es *EmailService) SetFrom(msg *gomail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.SetHeader("From", from)
}

func (es *EmailService) Send(email Email) error {
	// Email message headers
	msg := gomail.NewMessage()
	es.SetFrom(msg, email)
	msg.SetHeader("To", email.To)
	msg.SetHeader("Subject", email.Subject)

	// Email message body
	switch {
	case email.Plaintext != "" && email.HTML != "":
		msg.SetBody("text/plain", email.Plaintext)
		msg.AddAlternative("text/html", email.HTML)
	case email.Plaintext != "":
		msg.SetBody("text/plain", email.Plaintext)
	case email.HTML != "":
		msg.SetBody("text/html", email.HTML)
	}

	// Send the email
	if err := es.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (es *EmailService) ForgotPassword(to, resetURL string) error {
	email := Email{
		To: to,
		Subject: "Reset your password",
		Plaintext: "To reset your password, please visit the following link: " + resetURL,
		HTML: `<p>To reset your password, please visit the following link: <a href="` + resetURL + `">` + resetURL + `</a></p>`,
	}

	if err := es.Send(email); err != nil {
		return fmt.Errorf("forgot password email: %w", err)
	}
	return nil
}
