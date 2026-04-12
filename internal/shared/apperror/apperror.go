package apperror

import "errors"

// Sentinel errors for domain-level failures.
// Handlers map these to HTTP status codes — no HTTP knowledge leaks into domain/application.

var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("resource already exists")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("internal server error")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrTokenRevoked  = errors.New("token has been revoked")
	ErrWrongPassword = errors.New("wrong password")
	ErrInvalidTenant = errors.New("invalid tenant")
)

// AppError wraps a sentinel with a human-readable message for the client.
type AppError struct {
	Err     error
	Message string
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(sentinel error, message string) *AppError {
	return &AppError{Err: sentinel, Message: message}
}

// Is allows errors.Is to match the underlying sentinel.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
