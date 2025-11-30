package controllers

import "errors"

var (
	ErrInvalidFile						= errors.New("controllers: file has invalid content type or extension")
	ErrMissingRequiredFields	= errors.New("controllers: missing required fields")
	ErrPasswordsDontMatch			= errors.New("controllers: password and confirm password don't match")
	ErrUnauthorized						= errors.New("controllers: user is unauthorized to access resource")
)
