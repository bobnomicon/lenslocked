package models

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	// The default time that a password reset token is valid for.
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID int
	UserId int
	// Token only set when creating a new password reset. Will be empty otherwise, as only the hashed password reset token will be stored in the db.
	Token string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	BytesPerToken int // How many bytes to use when generating password reset token. If not set or is < MinBytesPerToken const (defined in session.go), it will be ignored and MinBytesPerToken used instead.
	Duration time.Duration // The amount of time that a password reset token is valid for. Defaults to DefaultResetDuration.
}

func (prs *PasswordResetService) Create(email string) (*PasswordReset, error) {
	// TODO!

	return nil, fmt.Errorf("TODO! Implement PasswordResetService.Create")
}

func (prs *PasswordResetService) Consume(token string) (*User, error) {
	// TODO!

	return nil, fmt.Errorf("TODO! Implement PasswordResetService.Consume")
}
