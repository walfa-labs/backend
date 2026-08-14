package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/walfa-labs/backend/internal/domain"
)

// SuccessEnvelope wraps a successful response per the design doc §4.5.3.
type SuccessEnvelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// Meta carries pagination metadata.
type Meta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"perPage"`
	Total   int64 `json:"total"`
}

// ErrorEnvelope wraps an error response per the design doc §4.5.3.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the error detail payload.
type ErrorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []domain.FieldError `json:"details,omitempty"`
}

// OK sends a 200 with data in the success envelope.
func OK(c fiber.Ctx, data any) error {
	return c.JSON(SuccessEnvelope{Data: data})
}

// OKWithMeta sends a 200 with data + pagination meta.
func OKWithMeta(c fiber.Ctx, data any, meta *Meta) error {
	return c.JSON(SuccessEnvelope{Data: data, Meta: meta})
}

// Created sends a 201 with data and a Location header.
func Created(c fiber.Ctx, location string, data any) error {
	c.Set("Location", location)
	c.Set("Cache-Control", "no-store")
	return c.Status(fiber.StatusCreated).JSON(SuccessEnvelope{Data: data})
}

// NoContent sends a 204 with no body.
func NoContent(c fiber.Ctx) error {
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// PublicCacheHeaders sets the RMM Level 2 cache headers on public read endpoints (§4.5.4).
func PublicCacheHeaders(c fiber.Ctx, etag string) {
	c.Set("Cache-Control", "public, max-age=60, s-maxage=300, stale-while-revalidate=600")
	c.Set("ETag", `"`+etag+`"`)
}

// NoStoreHeaders sets no-store on mutating/auth endpoints.
func NoStoreHeaders(c fiber.Ctx) {
	c.Set("Cache-Control", "no-store")
}
