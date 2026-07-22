package domain

import (
	"errors"
	"fmt"
)

// Domain errors. Sentinels + typed wrappers, mapped to HTTP by the error middleware.
var (
	// ErrNotFound is returned when a resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned on uniqueness violations (slug dup, etc.).
	ErrConflict = errors.New("conflict")
	// ErrValidation is returned when a domain invariant is violated.
	ErrValidation = errors.New("validation failed")
	// ErrUnauthorized is returned when authentication fails or is missing.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when an authenticated user lacks permission.
	ErrForbidden = errors.New("forbidden")
)

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// ValidationError carries per-field details for 400 responses.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d field error(s)", len(e.Fields))
}

func (e *ValidationError) Unwrap() error { return ErrValidation }

// NewValidationError builds a ValidationError from field/issue pairs.
// Each pair must be (field string, issue string); panics if odd.
func NewValidationError(fields ...string) *ValidationError {
	if len(fields)%2 != 0 {
		panic("domain: NewValidationError requires field/issue pairs")
	}
	fe := make([]FieldError, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		fe = append(fe, FieldError{Field: fields[i], Issue: fields[i+1]})
	}
	return &ValidationError{Fields: fe}
}
