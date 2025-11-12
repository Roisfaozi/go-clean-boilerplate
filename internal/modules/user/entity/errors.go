package entity

import "errors"

var (
	// ErrUserNotFound returned when user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrContactNotFound returned when contact is not found
	ErrContactNotFound = errors.New("contact not found")

	// ErrInvalidCredentials returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserAlreadyExists returned when trying to create existing user
	ErrUserAlreadyExists = errors.New("user already exists")
)
