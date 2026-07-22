package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type AuthHandler struct {
	svc         port.AuthService
	refreshTTLHours int
}

func NewAuthHandler(svc port.AuthService, refreshTTLHours int) *AuthHandler {
	return &AuthHandler{svc: svc, refreshTTLHours: refreshTTLHours}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	tokens, err := h.svc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return domain.ErrUnauthorized
		}
		return err
	}
	// Set refresh token as httpOnly cookie (§4.6, §6.6).
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   h.refreshTTLHours * 3600,
		Path:     "/api/v1/auth",
	})
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: loginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	// Prefer httpOnly cookie, fall back to body.
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		var req refreshRequest
		if err := c.Bind().Body(&req); err != nil {
			return domain.NewValidationError("body", "invalid JSON")
		}
		refreshToken = req.RefreshToken
	}
	if refreshToken == "" {
		return domain.ErrUnauthorized
	}
	accessToken, err := h.svc.Refresh(c.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return domain.ErrUnauthorized
		}
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: refreshResponse{AccessToken: accessToken}})
}
