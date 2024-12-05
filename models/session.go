package models

import "database/sql"

type Session struct {
	ID int
	UserID int
	// Token only set when creating a new session. Will be empty otherwise, as only the hashed session token will be stored in the db.
	Token string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
}

// Creates a new session for the provided user. The session token will be returned in the Token field of the Session type, but only the hashed session token will be stored in the db.
func (ss *SessionService) Create(userID int) (*Session, error) {
	// TODO
	return nil, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// TODO
	return nil, nil
}
