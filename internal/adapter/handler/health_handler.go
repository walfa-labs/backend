package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /api/v1/health.
func (h *HealthHandler) Health(c fiber.Ctx) error {
	status := "ok"
	dbStatus := "up"
	if err := h.db.Ping(c.Context()); err != nil {
		status = "degraded"
		dbStatus = "down"
	}
	return c.JSON(fiber.Map{
		"status": status,
		"db":     dbStatus,
	})
}
