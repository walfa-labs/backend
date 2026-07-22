package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// localsKey is the unexported type used for Locals keys, preventing collisions
// with keys set by other packages.
type localsKey int

const (
	// requestIDKey stores the request ID string in Locals.
	requestIDKey localsKey = iota
	// adminKey stores the authenticated admin username in Locals.
	adminKey
)

// RequestIDHeader is the HTTP header carrying the request correlation ID.
const RequestIDHeader = "X-Request-Id"

// RequestID generates a UUID request ID for each request. If the inbound
// request already carries an X-Request-Id header it is reused; otherwise a
// fresh UUIDv4 is generated. The ID is stored in Locals and echoed back on
// the response header so downstream services and clients can correlate logs.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Locals(requestIDKey, id)
		c.Set(RequestIDHeader, id)
		return c.Next()
	}
}

// RequestIDFromContext returns the request ID stored in Locals, or the empty
// string if none is set.
func RequestIDFromContext(c fiber.Ctx) string {
	if v, ok := c.Locals(requestIDKey).(string); ok {
		return v
	}
	return ""
}
