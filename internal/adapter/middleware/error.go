package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/walfa-labs/backend/internal/domain"
)

// Error envelope codes (§4.5.3 of the design doc).
const (
	CodeNotFound      = "NOT_FOUND"
	CodeConflict      = "CONFLICT"
	CodeValidation    = "VALIDATION_FAILED"
	CodeUnauthorized  = "UNAUTHORIZED"
	CodeForbidden     = "FORBIDDEN"
	CodeInternalError = "INTERNAL_ERROR"
)

// ErrorDetail is a single field-level validation issue in the error envelope.
type ErrorDetail struct {
	Field string `json:"field,omitempty"`
	Issue string `json:"issue,omitempty"`
}

// ErrorBody is the value of the "error" key in the JSON response.
type ErrorBody struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
}

// ErrorEnvelope is the top-level error response body:
//
//	{ "error": { "code": "...", "message": "...", "details": [...] } }
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorHandler maps domain errors to HTTP status codes and renders the
// standard error envelope. It is intended to be installed as fiber.Config.
// ErrorHandler so that any error returned from a handler is processed here.
func ErrorHandler(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// If the handler already wrote a response, Fiber wraps it in
	// fiber.ErrInternalServerError-style errors. Honor an explicit fiber error
	// status when present so e.g. c.SendStatus(404) isn't overridden.
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeInternalError,
				Message:   fiberErr.Message,
				RequestID: RequestIDFromContext(c),
			},
		})
	}

	// Typed validation error: extract per-field details.
	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		details := make([]ErrorDetail, len(valErr.Fields))
		for i, f := range valErr.Fields {
			details[i] = ErrorDetail{Field: f.Field, Issue: f.Issue}
		}
		return c.Status(fiber.StatusBadRequest).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeValidation,
				Message:   valErr.Error(),
				Details:   details,
				RequestID: RequestIDFromContext(c),
			},
		})
	}

	// Sentinel domain errors.
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeNotFound,
				Message:   "resource not found",
				RequestID: RequestIDFromContext(c),
			},
		})
	case errors.Is(err, domain.ErrConflict):
		return c.Status(fiber.StatusConflict).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeConflict,
				Message:   "conflict with existing resource",
				RequestID: RequestIDFromContext(c),
			},
		})
	case errors.Is(err, domain.ErrUnauthorized):
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeUnauthorized,
				Message:   "authentication required or failed",
				RequestID: RequestIDFromContext(c),
			},
		})
	case errors.Is(err, domain.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeForbidden,
				Message:   "insufficient permissions",
				RequestID: RequestIDFromContext(c),
			},
		})
	case errors.Is(err, domain.ErrValidation):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorEnvelope{
			Error: ErrorBody{
				Code:      CodeValidation,
				Message:   err.Error(),
				RequestID: RequestIDFromContext(c),
			},
		})
	}

	// Fallback: any unmapped error is a 500.
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorEnvelope{
		Error: ErrorBody{
			Code:      CodeInternalError,
			Message:   "an internal server error occurred",
			RequestID: RequestIDFromContext(c),
		},
	})
}
