package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// ProjectHandler handles project request endpoints.
type ProjectHandler struct {
	svc port.ProjectService
}

// NewProjectHandler constructs the project HTTP handlers bound to the project service.
func NewProjectHandler(svc port.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List returns published projects (public endpoint).
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

// GetBySlug returns one public project by slug.
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

// AdminList returns all projects including drafts.
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

// AdminGet returns any project by id.
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

// Create validates input and persists a new project.
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

// Update validates input and persists changes to an existing project.
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

// Delete removes the project with the given id.
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
