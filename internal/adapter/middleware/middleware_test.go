package middleware_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/adapter/middleware"
	"github.com/walfa-labs/backend/internal/config"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/platform"
)

type mockConfig struct {
	secret string
}

func (m mockConfig) JWTSecret() string {
	return m.secret
}

func TestRequestIDMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/test", func(c fiber.Ctx) error {
		reqID := middleware.RequestIDFromContext(c)
		return c.SendString(reqID)
	})

	t.Run("generates UUID when header is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		reqID := string(body)
		if _, err := uuid.Parse(reqID); err != nil {
			t.Fatalf("expected valid UUID, got '%s'", reqID)
		}

		headerID := resp.Header.Get(middleware.RequestIDHeader)
		if headerID != reqID {
			t.Errorf("header ID %s does not match body ID %s", headerID, reqID)
		}
	})

	t.Run("preserves existing X-Request-Id header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, "custom-trace-id-123")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "custom-trace-id-123" {
			t.Errorf("expected 'custom-trace-id-123', got '%s'", string(body))
		}
		if resp.Header.Get(middleware.RequestIDHeader) != "custom-trace-id-123" {
			t.Errorf("expected header 'custom-trace-id-123', got '%s'", resp.Header.Get(middleware.RequestIDHeader))
		}
	})
}

func TestErrorHandler(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	app.Use(middleware.RequestID())

	app.Get("/not-found", func(c fiber.Ctx) error {
		return domain.ErrNotFound
	})
	app.Get("/conflict", func(c fiber.Ctx) error {
		return domain.ErrConflict
	})
	app.Get("/unauthorized", func(c fiber.Ctx) error {
		return domain.ErrUnauthorized
	})
	app.Get("/forbidden", func(c fiber.Ctx) error {
		return domain.ErrForbidden
	})
	app.Get("/validation", func(c fiber.Ctx) error {
		return domain.NewValidationError("title", "is required", "email", "invalid format")
	})
	app.Get("/fiber-error", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusPaymentRequired, "payment required")
	})
	app.Get("/internal", func(c fiber.Ctx) error {
		return errors.New("database connection refused")
	})

	tests := []struct {
		name         string
		path         string
		expectedCode int
		expectedType string
	}{
		{"NotFound", "/not-found", http.StatusNotFound, middleware.CodeNotFound},
		{"Conflict", "/conflict", http.StatusConflict, middleware.CodeConflict},
		{"Unauthorized", "/unauthorized", http.StatusUnauthorized, middleware.CodeUnauthorized},
		{"Forbidden", "/forbidden", http.StatusForbidden, middleware.CodeForbidden},
		{"Validation", "/validation", http.StatusBadRequest, middleware.CodeValidation},
		{"FiberError", "/fiber-error", http.StatusPaymentRequired, middleware.CodeInternalError},
		{"InternalError", "/internal", http.StatusInternalServerError, middleware.CodeInternalError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, resp.StatusCode)
			}

			var env middleware.ErrorEnvelope
			if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
				t.Fatalf("failed to decode error envelope: %v", err)
			}

			if env.Error.Code != tc.expectedType {
				t.Errorf("expected error code '%s', got '%s'", tc.expectedType, env.Error.Code)
			}

			if tc.name == "Validation" {
				if len(env.Error.Details) != 2 {
					t.Errorf("expected 2 error details, got %d", len(env.Error.Details))
				}
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := "my-test-secret-key"
	cfg := mockConfig{secret: jwtSecret}
	logger := platform.NewLogger("development")

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	app.Use(middleware.RequestID())
	app.Use(middleware.Auth(cfg, logger))

	app.Get("/public", func(c fiber.Ctx) error {
		return c.SendString("public ok")
	})

	adminGroup := app.Group("/admin", middleware.RequireAdmin())
	adminGroup.Get("/dashboard", func(c fiber.Ctx) error {
		claims, ok := middleware.AdminFromContext(c)
		if !ok {
			return domain.ErrUnauthorized
		}
		return c.SendString("welcome " + claims.Username)
	})

	generateToken := func(username string, secret string, expired bool) string {
		exp := time.Now().Add(1 * time.Hour)
		if expired {
			exp = time.Now().Add(-1 * time.Hour)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.AdminClaims{
			AdminID:  uuid.New(),
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   username,
				ExpiresAt: jwt.NewNumericDate(exp),
			},
		})
		s, _ := token.SignedString([]byte(secret))
		return s
	}

	t.Run("public endpoint accessible without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("admin endpoint returns 401 without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("admin endpoint succeeds with valid token", func(t *testing.T) {
		validToken := generateToken("walfa_admin", jwtSecret, false)
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "welcome walfa_admin" {
			t.Errorf("expected 'welcome walfa_admin', got '%s'", string(body))
		}
	})

	t.Run("admin endpoint returns 401 with invalid signature token", func(t *testing.T) {
		wrongToken := generateToken("walfa_admin", "wrong-secret", false)
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+wrongToken)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("admin endpoint returns 401 with expired token", func(t *testing.T) {
		expiredToken := generateToken("walfa_admin", jwtSecret, true)
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"http://localhost:3000", "https://walfa.dev"},
	}

	app := fiber.New()
	app.Use(middleware.CORS(cfg))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://walfa.dev")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "https://walfa.dev" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://walfa.dev', got '%s'", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestRecoverMiddleware(t *testing.T) {
	logger := platform.NewLogger("development")
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Use(middleware.Recover(logger))

	app.Get("/panic", func(c fiber.Ctx) error {
		panic("something went critically wrong!")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 status code, got %d", resp.StatusCode)
	}

	var env middleware.ErrorEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode error envelope: %v", err)
	}

	if env.Error.Code != middleware.CodeInternalError {
		t.Errorf("expected code '%s', got '%s'", middleware.CodeInternalError, env.Error.Code)
	}
}
