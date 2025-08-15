package errors

import "errors"

// ServiceError represents a service-level error
type ServiceError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	return e.Message
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return err != nil && (errors.Is(err, ErrNotFound) || err.Error() == "not found")
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) error {
	if message == "" {
		return ErrNotFound
	}
	return errors.New(message)
}

// Common errors
var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
	ErrInternal = errors.New("internal error")
)
