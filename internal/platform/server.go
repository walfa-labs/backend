package platform

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gookit/slog"

	"github.com/walfa-labs/backend/internal/adapter/middleware"
	"github.com/walfa-labs/backend/internal/config"
)

// NewServer creates a configured *fiber.App. It applies the Sonic JSON
// encoder/decoder from FiberConfig, installs the StructValidator for request
// binding, and returns the bare app. Route and middleware registration is the
// responsibility of the adapter/router package — this factory only wires
// cross-cutting configuration that must exist before any route is added.
func NewServer(cfg *config.Config, logger *slog.Logger) *fiber.App {
	fcfg := FiberConfig()

	// Install the validator so Bind() calls validate automatically.
	v := NewValidator()
	fcfg.StructValidator = NewStructValidator(v, logger)
	fcfg.ErrorHandler = middleware.ErrorHandler

	app := fiber.New(fcfg)
	return app
}
