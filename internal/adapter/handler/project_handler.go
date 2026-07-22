package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

type ProjectHandler struct {
	svc port.ProjectService
}

func NewProjectHandler(svc port.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) List(c fiber.Ctx) error {
	var featured *bool
	if q := c.Query("featured"); q != "" {
		f := q == "1" || q == "true"
		featured = &f
	}
	projects, err := h.svc.ListPublished(c.Context(), featured)
	if err != nil {
		return err
	}
	results := make([]ProjectResponse, 0, len(projects))
	for i := range projects {
		results = append(results, toProjectResponse(&projects[i]))
	}
	return OK(c, results)
}

func (h *ProjectHandler) GetBySlug(c fiber.Ctx) error {
	p, err := h.svc.GetPublishedBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return OK(c, toProjectResponse(p))
}

func (h *ProjectHandler) AdminList(c fiber.Ctx) error {
	projects, err := h.svc.ListAll(c.Context())
	if err != nil {
		return err
	}
	results := make([]ProjectResponse, 0, len(projects))
	for i := range projects {
		results = append(results, toProjectResponse(&projects[i]))
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: results})
}

func (h *ProjectHandler) AdminGet(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	p, err := h.svc.Get(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toProjectResponse(p)})
}

func (h *ProjectHandler) Create(c fiber.Ctx) error {
	var req projectRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	input, err := req.toInput()
	if err != nil {
		return err
	}
	p, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return err
	}
	return Created(c, "/api/v1/admin/projects/"+p.ID.String(), toProjectResponse(p))
}

func (h *ProjectHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	var req projectRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	input, err := req.toInput()
	if err != nil {
		return err
	}
	p, err := h.svc.Update(c.Context(), id, input)
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toProjectResponse(p)})
}

func (h *ProjectHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return NoContent(c)
}

