package exception

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInternalServer      = errors.New("internal server error")
	ErrConflict            = errors.New("data already exists")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrValidationError     = errors.New("validation error")
	ErrTooManyRequests     = errors.New("to many requests")
)
