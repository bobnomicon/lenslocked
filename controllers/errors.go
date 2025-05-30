package controllers

import "errors"

var (
	ErrMissingRequiredFields	= errors.New("controllers: missing required fields")
	ErrPasswordsDontMatch			= errors.New("controllers: password and confirm password don't match")
)
