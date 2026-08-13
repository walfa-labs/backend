package handler

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"
)

// HealthHandler reports liveness, including connectivity to both Oracle
// databases (ATP = operational store, ADW = analytics store).
type HealthHandler struct {
	atp *sql.DB
	adw *sql.DB
}

func NewHealthHandler(atp, adw *sql.DB) *HealthHandler {
	return &HealthHandler{atp: atp, adw: adw}
}

// Health handles GET /api/v1/health.
func (h *HealthHandler) Health(c fiber.Ctx) error {
	status := "ok"
	dbStatus := "up"
	if err := h.atp.PingContext(c.Context()); err != nil {
		status = "degraded"
		dbStatus = "down"
	}
	if err := h.adw.PingContext(c.Context()); err != nil {
		status = "degraded"
		dbStatus = "down"
	}
	return c.JSON(fiber.Map{
		"status": status,
		"db":     dbStatus,
	})
}
