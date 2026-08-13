package middleware

import "github.com/gofiber/fiber/v3"

// SecurityHeaders sets baseline HTTP security headers on all responses.
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Cross-Origin-Resource-Policy", "cross-origin")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	}
}
