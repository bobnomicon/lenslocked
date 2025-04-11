package models

import "errors"

var (
	ErrEmailDoesNotExist	= errors.New("models: email does not exist")
	ErrEmailTaken					= errors.New("models: email address already taken")
	ErrInvalidCredentials = errors.New("models: invalid authentication credentials")
	ErrNotFound						= errors.New("models: resource could not be found")
	ErrTokenExpired				= errors.New("modles: password reset token expired")
	ErrTokenInvalid				= errors.New("models: password reset token does not exist")
)
