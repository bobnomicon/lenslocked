package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/operas-logicas/lenslocked/rand"
)

const (
	// The default time that a password reset token is valid for.
	DefaultResetDuration = 1 * time.Hour
)

var (
	ErrEmailDoesNotExist = errors.New("models: email does not exist")
	ErrTokenInvalid = errors.New("models: password reset token does not exist")
	ErrTokenExpired = errors.New("modles: password reset token expired")
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

func (prs *PasswordResetService) delete(id int) error {
	_, err := prs.DB.Exec(`
		DELETE FROM password_resets
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
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
		if errors.Is(err, sql.ErrNoRows) {
				// Email does not exist
				return nil, ErrEmailDoesNotExist
		}

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
	// Hash password reset token
	tokenHash := prs.hash(token)

	var user User
	var passwordReset PasswordReset

	// Query db for user with password reset token hash
	row := prs.DB.QueryRow(`
		SELECT
			users.id,
			users.email,
			users.password_hash,
			password_resets.id,
			password_resets.expires_at
		FROM users
			JOIN password_resets ON password_resets.user_id = users.id
		WHERE password_resets.token_hash = $1
	`, tokenHash)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&passwordReset.ID,
		&passwordReset.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Token does not exist (invalid token)
			return nil, ErrTokenInvalid
		}

		return nil, fmt.Errorf("consume: %w", err)
	}

	// Check token is still valid (not expired)
	if time.Now().After(passwordReset.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Password reset token is valid, so delete it from db
	err = prs.delete(passwordReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return &user, nil
}
