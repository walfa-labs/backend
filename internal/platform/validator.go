package platform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gookit/slog"

	"github.com/walfa-labs/backend/internal/domain"
)

// NewValidator returns a validator.Validate configured with sensible defaults.
// JSON tags are used for field names so error fields match request payloads.
func NewValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v
}

// StructValidator adapts go-playground/validator to Fiber v3's
// bind.StructValidator interface (Validate(out any) error). Validation
// failures are returned as a *domain.ValidationError carrying per-field
// details, which the error middleware maps to a 400 response.
type StructValidator struct {
	validate *validator.Validate
	logger   *slog.Logger
}

// NewStructValidator returns a StructValidator wired to the given logger.
func NewStructValidator(v *validator.Validate, logger *slog.Logger) *StructValidator {
	return &StructValidator{validate: v, logger: logger}
}

// Validate runs struct validation on out and returns a *domain.ValidationError
// when fields fail, so the central error handler can produce 400 bodies with
// field-level details.
func (sv *StructValidator) Validate(out any) error {
	if err := sv.validate.Struct(out); err != nil {
		// Not a validation error (e.g. nil/non-struct passed) — surface as-is.
		var verrs validator.ValidationErrors
		if !AsValidationErrors(err, &verrs) {
			if sv.logger != nil {
				sv.logger.Errorf("validator: unexpected error: %v", err)
			}
			return err
		}
		return translateValidationErrors(verrs)
	}
	return nil
}

// AsValidationErrors reports whether err wraps validator.ValidationErrors.
func AsValidationErrors(err error, target *validator.ValidationErrors) bool {
	return errors.As(err, target)
}

// translateValidationErrors converts validator.FieldError slices into a
// domain.ValidationError with human-readable issue strings.
func translateValidationErrors(verrs validator.ValidationErrors) *domain.ValidationError {
	fields := make([]string, 0, len(verrs)*2)
	for _, fe := range verrs {
		fields = append(fields, fe.Field(), issueFor(fe))
	}
	return domain.NewValidationError(fields...)
}

// issueFor produces a concise description of a single validation failure.
func issueFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "slug":
		return "must contain only lowercase letters, numbers, and hyphens"
	default:
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}
