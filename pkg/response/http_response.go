package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

// WriteJSON encodes data as JSON into http.ResponseWriter with the given status code.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a standard successful response.
func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	WriteJSON(w, statusCode, WebResponseSuccess[any]{
		Data: data,
	})
}

// WriteSuccessWithPaging writes a standard successful response with paging metadata.
func WriteSuccessWithPaging(w http.ResponseWriter, data any, paging *PageMetadata) {
	WriteJSON(w, http.StatusOK, WebResponseSuccess[any]{
		Data:   data,
		Paging: paging,
	})
}

// WriteError writes a standard error response.
func WriteError(w http.ResponseWriter, statusCode int, err error, msg string) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	WriteJSON(w, statusCode, WebResponseError[any]{
		Error:   errorMsg,
		Message: msg,
	})
}

// WriteHTTPError maps domain exception sentinel errors to HTTP status codes.
func WriteHTTPError(w http.ResponseWriter, err error, message string) {
	switch {
	case errors.Is(err, exception.ErrBadRequest):
		WriteError(w, http.StatusBadRequest, err, message)
	case errors.Is(err, exception.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, err, message)
	case errors.Is(err, exception.ErrForbidden):
		WriteError(w, http.StatusForbidden, err, message)
	case errors.Is(err, exception.ErrNotFound):
		WriteError(w, http.StatusNotFound, err, message)
	case errors.Is(err, exception.ErrConflict):
		WriteError(w, http.StatusConflict, err, message)
	case errors.Is(err, exception.ErrValidationError), errors.Is(err, exception.ErrUnprocessableEntity):
		WriteError(w, http.StatusUnprocessableEntity, err, message)
	case errors.Is(err, exception.ErrTooManyRequests):
		WriteError(w, http.StatusTooManyRequests, err, message)
	default:
		WriteError(w, http.StatusInternalServerError, err, message)
	}
}

// DecodeJSON decodes request JSON body into target struct, enforcing max size and no trailing tokens.
func DecodeJSON(r *http.Request, target any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1MB default
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(target); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("extra data after JSON payload")
	}
	return nil
}

// ReadBody reads full body bytes up to maxBytes.
func ReadBody(r *http.Request, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	return io.ReadAll(io.LimitReader(r.Body, maxBytes))
}
