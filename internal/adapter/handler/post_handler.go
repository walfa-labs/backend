package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// PostHandler handles post request endpoints.
type PostHandler struct {
	svc port.PostService
}

// NewPostHandler constructs the post HTTP handlers bound to the post service.
func NewPostHandler(svc port.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// List returns published posts (public endpoint).
func (h *PostHandler) List(c fiber.Ctx) error {
	q := c.Queries()
	page, _ := strconv.Atoi(q["page"])
	perPage, _ := strconv.Atoi(q["perPage"])

	posts, total, err := h.svc.ListPublished(c.Context(), port.PostFilter{
		Tag:     q["tag"],
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return err
	}
	results := make([]PostSummaryResponse, 0, len(posts))
	for _, p := range posts {
		results = append(results, toPostSummaryResponse(p))
	}
	return OKWithMeta(c, results, &Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GetBySlug returns one public post by slug.
func (h *PostHandler) GetBySlug(c fiber.Ctx) error {
	p, err := h.svc.GetPublishedBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	etag := p.ETag()
	if c.Get("If-None-Match") == `"`+etag+`"` {
		return c.SendStatus(fiber.StatusNotModified)
	}
	PublicCacheHeaders(c, etag)
	return c.JSON(SuccessEnvelope{Data: toPostResponse(p)})
}

// AdminList returns all posts including drafts.
func (h *PostHandler) AdminList(c fiber.Ctx) error {
	posts, err := h.svc.ListAll(c.Context())
	if err != nil {
		return err
	}
	results := make([]PostResponse, 0, len(posts))
	for i := range posts {
		results = append(results, toPostResponse(&posts[i]))
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: results})
}

// AdminGet returns any post by id.
func (h *PostHandler) AdminGet(c fiber.Ctx) error {
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
	return c.JSON(SuccessEnvelope{Data: toPostResponse(p)})
}

// Create validates input and persists a new post.
func (h *PostHandler) Create(c fiber.Ctx) error {
	var req postRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	p, err := h.svc.Create(c.Context(), req.toInput())
	if err != nil {
		return err
	}
	return Created(c, "/api/v1/admin/blog/posts/"+p.ID.String(), toPostResponse(p))
}

// Update validates input and persists changes to an existing post.
func (h *PostHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	var req postRequest
	if err := c.Bind().Body(&req); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	p, err := h.svc.Update(c.Context(), id, req.toInput())
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toPostResponse(p)})
}

// Delete removes the post with the given id.
func (h *PostHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return NoContent(c)
}

// SetStatus transitions a post between draft and published.
func (h *PostHandler) SetStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.NewValidationError("id", "invalid UUID")
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return domain.NewValidationError("body", "invalid JSON")
	}
	status := domain.ContentStatus(body.Status)
	p, err := h.svc.SetStatus(c.Context(), id, status)
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return c.JSON(SuccessEnvelope{Data: toPostResponse(p)})
}
