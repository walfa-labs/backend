package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gookit/slog"

	"github.com/walfa-labs/backend/internal/domain"
)

// AdminClaims is the JWT claims payload for an authenticated admin user.
// It embeds jwt.RegisteredClaims for standard fields (sub, exp, iat, jti).
type AdminClaims struct {
	AdminID   uuid.UUID `json:"admin_id"`
	Username  string    `json:"username"`
	jwt.RegisteredClaims
}

// AuthScheme is the expected Authorization header scheme prefix.
const AuthScheme = "Bearer"

// AdminFromContext reports whether the current request was made by an
// authenticated admin, returning the claims when present.
func AdminFromContext(c fiber.Ctx) (*AdminClaims, bool) {
	if v, ok := c.Locals(adminKey).(*AdminClaims); ok && v != nil {
		return v, true
	}
	return nil, false
}

// RequireAdmin is a guard middleware for admin-only routes. It returns 401
// when no admin identity is present (Auth middleware did not run or failed).
// Place it after Auth in the chain.
func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, ok := AdminFromContext(c); ok {
			return c.Next()
		}
		return domain.ErrUnauthorized
	}
}

// Auth returns JWT authentication middleware. It extracts the Bearer token
// from the Authorization header, verifies the signature and expiry against
// cfg.JWTSecret, and populates Locals with the admin identity (AdminClaims).
// Requests without a valid token are allowed to continue (public routes);
// route-level access control is enforced by RequireAdmin. This lets the same
// middleware run globally without rejecting unauthenticated public reads.
func Auth(cfg jwtKeyReader, logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		raw := extractBearerToken(c)
		if raw == "" {
			return c.Next()
		}

		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(cfg.JWTSecret()), nil
		})

		if err != nil || !token.Valid {
			if logger != nil {
				logger.Debugf("auth: token verification failed: %v", err)
			}
			return c.Next()
		}

		c.Locals(adminKey, claims)
		return c.Next()
	}
}

// jwtKeyReader is a minimal interface over config.Config so the middleware
// can read the JWT secret without importing the config package directly
// (keeping the test surface small). *config.Config satisfies this.
type jwtKeyReader interface {
	JWTSecret() string
}

// extractBearerToken parses the Authorization header and returns the raw
// token string, or "" when the header is absent or not a Bearer token.
func extractBearerToken(c fiber.Ctx) string {
	auth := c.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], AuthScheme) {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
