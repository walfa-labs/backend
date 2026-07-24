package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type ProfileHandler struct {
	svc port.ProfileService
}

func NewProfileHandler(svc port.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

// Get handles GET /api/v1/profile (public).
func (h *ProfileHandler) Get(c fiber.Ctx) error {
	profile, err := h.svc.Get(c.Context())
	if err != nil {
		return err
	}
	return OK(c, toProfileResponse(profile))
}

// AdminGet handles GET /api/v1/admin/profile (admin).
func (h *ProfileHandler) AdminGet(c fiber.Ctx) error {
	profile, err := h.svc.Get(c.Context())
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toProfileResponse(profile)})
}

// Update handles PUT /api/v1/admin/profile (admin).
func (h *ProfileHandler) Update(c fiber.Ctx) error {
	var req profileRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	profile, err := h.svc.Update(c.Context(), req.toInput())
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toProfileResponse(profile)})
}
