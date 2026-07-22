package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/walfa-labs/backend/internal/port"
)

type StatsHandler struct {
	svc     port.StatsService
	tagRepo port.TagRepo
}

func NewStatsHandler(svc port.StatsService, tagRepo port.TagRepo) *StatsHandler {
	return &StatsHandler{svc: svc, tagRepo: tagRepo}
}

// Summary handles GET /api/v1/stats/summary.
func (h *StatsHandler) Summary(c fiber.Ctx) error {
	summary, err := h.svc.Summary(c.Context())
	if err != nil {
		return err
	}
	PublicCacheHeaders(c, "summary")
	return OK(c, toStatsSummaryResponse(summary))
}

// ViewsTimeSeries handles GET /api/v1/admin/stats/views.
func (h *StatsHandler) ViewsTimeSeries(c fiber.Ctx) error {
	q := c.Queries()
	var from, to time.Time
	if v := q["from"]; v != "" {
		from, _ = time.Parse(time.RFC3339, v)
	}
	if v := q["to"]; v != "" {
		to, _ = time.Parse(time.RFC3339, v)
	}
	bucket := q["bucket"]
	if bucket == "" {
		bucket = "day"
	}
	buckets, err := h.svc.ViewsTimeSeries(c.Context(), from, to, bucket)
	if err != nil {
		return err
	}
	results := make([]map[string]any, 0, len(buckets))
	for _, b := range buckets {
		results = append(results, map[string]any{
			"bucket": b.Bucket.Format(time.RFC3339),
			"views":  b.Views,
		})
	}
	NoStoreHeaders(c)
	return OK(c, results)
}

// TopPosts handles GET /api/v1/admin/stats/top-posts.
func (h *StatsHandler) TopPosts(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit"))
	top, err := h.svc.TopPosts(c.Context(), limit)
	if err != nil {
		return err
	}
	results := make([]map[string]any, 0, len(top))
	for _, p := range top {
		results = append(results, map[string]any{
			"id":    p.ID.String(),
			"slug":  p.Slug,
			"title": p.Title,
			"views": p.Views,
		})
	}
	NoStoreHeaders(c)
	return OK(c, results)
}

// Tags handles GET /api/v1/tags.
func (h *StatsHandler) Tags(c fiber.Ctx) error {
	tags, err := h.tagRepo.List(c.Context())
	if err != nil {
		return err
	}
	results := make([]TagResponse, 0, len(tags))
	for _, t := range tags {
		results = append(results, TagResponse{ID: t.ID.String(), Name: t.Name, Slug: t.Slug})
	}
	PublicCacheHeaders(c, "tags")
	return OK(c, results)
}
