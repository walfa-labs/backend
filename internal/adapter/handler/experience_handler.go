package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type ExperienceHandler struct {
	svc port.ExperienceService
}

func NewExperienceHandler(svc port.ExperienceService) *ExperienceHandler {
	return &ExperienceHandler{svc: svc}
}

func (h *ExperienceHandler) List(c fiber.Ctx) error {
	experiences, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	results := make([]ExperienceResponse, 0, len(experiences))
	for i := range experiences {
		results = append(results, toExperienceResponse(&experiences[i]))
	}
	return OK(c, results)
}

func (h *ExperienceHandler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	e, err := h.svc.Get(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return OK(c, toExperienceResponse(e))
}

func (h *ExperienceHandler) Create(c fiber.Ctx) error {
	var req experienceRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	input, err := req.toInput()
	if err != nil {
		return err
	}
	e, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return err
	}
	return Created(c, "/api/v1/admin/experiences/"+e.ID.String(), toExperienceResponse(e))
}

func (h *ExperienceHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	var req experienceRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	input, err := req.toInput()
	if err != nil {
		return err
	}
	e, err := h.svc.Update(c.Context(), id, input)
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toExperienceResponse(e)})
}

func (h *ExperienceHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return NoContent(c)
}
