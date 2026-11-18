// Package apperror defines typed application errors that map cleanly to HTTP responses.
package apperror

import (
	"errors"
	"net/http"
)

// Error is a domain error carrying an HTTP status and a stable machine-readable code.
type Error struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func NotFound(entity string) *Error {
	return &Error{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: entity + " not found"}
}

func Validation(fields map[string]string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "VALIDATION_FAILED",
		Message: "Please correct the highlighted fields", Fields: fields}
}

func BadRequest(msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: msg}
}

func Conflict(msg string) *Error {
	return &Error{Status: http.StatusConflict, Code: "CONFLICT", Message: msg}
}

func Unavailable(msg string) *Error {
	return &Error{Status: http.StatusConflict, Code: "BOOK_UNAVAILABLE", Message: msg}
}

func Rule(msg string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "RULE_VIOLATION", Message: msg}
}

func Internal(err error) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR",
		Message: "Something went wrong on our side. Please try again."}
}

// From converts any error into an *Error, wrapping unknown errors as internal errors.
func From(err error) *Error {
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return Internal(err)
}
