package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/bobnomicon/lenslocked/rand"
)

const (
	// The minimum number of bytes to use for each session token.
	MinBytesPerToken = 32
)

type Session struct {
	ID int
	UserID int
	// Token only set when creating a new session. Will be empty otherwise, as only the hashed session token will be stored in the db.
	Token string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	BytesPerToken int // How many bytes to use when generating session token. If not set or is < MinBytesPerToken const, it will be ignored and MinBytesPerToken used instead.
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// base64 encode the data into a string
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

// Creates a new session for the provided user. The session token will be returned in the Token field of the Session type, but only the hashed session token will be stored in the db.
func (ss *SessionService) Create(userID int) (*Session, error) {
	// Create the session token
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Hash session token
	tokenHash := ss.hash(token)

	session := Session{
		UserID: userID,
		Token: token,
		TokenHash: tokenHash,
	}

	// Update session in db with hashed session token
	row := ss.DB.QueryRow(`
		INSERT INTO sessions (user_id, token_hash)
		VALUES ($1, $2) ON CONFLICT (user_id) DO
		UPDATE SET token_hash = $2
		RETURNING id;
	`, userID, tokenHash)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &session, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// Hash session token
	tokenHash := ss.hash(token)

	// Query db for user with session token hash
	var user User
	row := ss.DB.QueryRow(`
		SELECT
			users.id,
			users.email,
			users.password_hash
		FROM users
			JOIN sessions ON sessions.user_id = users.id
		WHERE sessions.token_hash = $1;
	`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	return &user, nil
}

func (ss *SessionService) Delete(token string) error {
	// Hash session token
	tokenHash := ss.hash(token)

	// Delete session from db
	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1;
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}
