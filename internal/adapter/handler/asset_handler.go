package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/port"
)

// AssetHandler handles asset upload, redirect, and delete requests.
type AssetHandler struct {
	svc port.AssetService
}

// NewAssetHandler constructs the asset HTTP handlers bound to the asset service.
func NewAssetHandler(svc port.AssetService) *AssetHandler {
	return &AssetHandler{svc: svc}
}

// Upload handles multipart file upload (POST /api/v1/admin/assets).
func (h *AssetHandler) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return domain.NewValidationError("file", "no file uploaded")
	}
	if file.Size > 10*1024*1024 {
		return domain.NewValidationError("file", "max size is 10MB")
	}

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	asset, err := h.svc.Upload(c.Context(), f, file.Header.Get("Content-Type"), file.Size)
	if err != nil {
		return err
	}
	NoStoreHeaders(c)
	return Created(c, "/api/v1/admin/assets/"+asset.Key, toAssetResponse(asset))
}

// Redirect handles GET /api/v1/assets/* — redirects to a presigned URL.
// The wildcard captures multi-segment keys like "images/<uuid>.png".
func (h *AssetHandler) Redirect(c fiber.Ctx) error {
	key := c.Params("*")
	if key == "" {
		return domain.ErrNotFound
	}
	url, err := h.svc.GetURL(c.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return c.Redirect().To(url)
}

// Delete handles DELETE /api/v1/admin/assets/*.
func (h *AssetHandler) Delete(c fiber.Ctx) error {
	key := c.Params("*")
	if key == "" {
		return domain.NewValidationError("key", "required")
	}
	if err := h.svc.Delete(c.Context(), key); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return NoContent(c)
}
