package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/operas-logicas/lenslocked/rand"
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

func (prs *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// base64 encode the data into a string
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

// Creates a new password reset token for the provided user. The token will be returned in the Token field of the PasswordReset type, but only the hashed token will be stored in the db.
func (prs *PasswordResetService) Create(email string) (*PasswordReset, error) {
	email = strings.ToLower(email)

	// Get user ID
	var userID int
	row := prs.DB.QueryRow(`
		SELECT id
		FROM users WHERE email = $1
	`, email)
	err := row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create password reset: %w", err)
	}

	// Create password reset token
	bytesPerToken := prs.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create password reset: %w", err)
	}

	// Hash the token
	tokenHash := prs.hash(token)

	// Set expiration of the token
	duration := prs.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	expiresAt := time.Now().Add(duration)

	passwordReset := PasswordReset{
		UserId: userID,
		Token: token,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	// Update password reset in db with hashed token and expiration
	row = prs.DB.QueryRow(`
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
		UPDATE SET token_hash = $2, expires_at = $3
		RETURNING id;
	`, userID, tokenHash, expiresAt)
	err = row.Scan(&passwordReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &passwordReset, nil
}

func (prs *PasswordResetService) Consume(token string) (*User, error) {
	// TODO!

	return nil, fmt.Errorf("TODO! Implement PasswordResetService.Consume")
}
