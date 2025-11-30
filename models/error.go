package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
)

var (
	ErrEmailDoesNotExist	= errors.New("models: email does not exist")
	ErrEmailTaken					= errors.New("models: email address already taken")
	ErrInvalidCredentials = errors.New("models: invalid authentication credentials")
	ErrNotFound						= errors.New("models: resource could not be found")
	ErrTokenExpired				= errors.New("modles: password reset token expired")
	ErrTokenInvalid				= errors.New("models: password reset token does not exist")
)

type FileError struct {
	Issue string
}

func (fe FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", fe.Issue)
}

// Checks if a file's content type is in the allowed types, returns a FileError if content type is invalid.
func checkContentType(r io.ReadSeeker, allowedTypes []string) error {
	// Read in 512 bytes of the file
	testBytes := make([]byte, 512)
	_, err := r.Read(testBytes)
	if err != nil {
		return fmt.Errorf("check content type: %w", err)
	}

	// Reset the reader
	_, err = r.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("check content type: %w", err)
	}

	// Check the content type of the file
	contentType := http.DetectContentType(testBytes)
	if slices.Contains(allowedTypes, contentType) {
		return nil
	}
	return FileError{
		Issue: fmt.Sprintf("invalid content type: %v", contentType),
	}
}

// Checks if a file's extension is in the allowed extensions, returns a FileError if extension is invalid.
func checkExtension(filename string, allowedExtensions []string) error {
	if !hasExtension(filename, allowedExtensions...) {
		return FileError{
			Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
		}
	}

	return nil
}
