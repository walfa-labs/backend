package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gookit/slog"
)

// Logger returns a request-logging middleware. Each request is logged with
// method, path, status code, latency, and the correlation request ID. The
// logger is injected so the same configured *slog.Logger is used throughout.
func Logger(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		reqID := RequestIDFromContext(c)

		logger.Infof(
			"%s %s %d %s req_id=%s",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			latency,
			reqID,
		)

		return err
	}
}
