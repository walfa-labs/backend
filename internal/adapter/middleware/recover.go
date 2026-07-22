package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/gookit/slog"
)

// Recover returns a panic-recovery middleware. When a handler panics, the
// stack trace is logged and a 500 JSON error response is returned so the
// client receives a structured error envelope instead of a dropped connection.
func Recover(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					logger.Errorf(
						"panic recovered: %v\n%s",
						r,
						debug.Stack(),
					)
				}

				err = c.Status(fiber.StatusInternalServerError).JSON(ErrorEnvelope{
					Error: ErrorBody{
						Code:    CodeInternalError,
						Message: "an internal server error occurred",
						Details: nil,
						RequestID: RequestIDFromContext(c),
					},
				})
			}
		}()

		return c.Next()
	}
}
