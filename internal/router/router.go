package router

import (
	"os"
	"time"

	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gookit/slog"
	"github.com/walfa-labs/backend/internal/adapter/handler"
	"github.com/walfa-labs/backend/internal/adapter/middleware"
	"github.com/walfa-labs/backend/internal/config"
)

// Deps bundles all handler + middleware dependencies for the router.
type Deps struct {
	Cfg            *config.Config
	Health         *handler.HealthHandler
	Experience     *handler.ExperienceHandler
	Project        *handler.ProjectHandler
	Post           *handler.PostHandler
	Auth           *handler.AuthHandler
	Asset          *handler.AssetHandler
	Stats          *handler.StatsHandler
	Profile        *handler.ProfileHandler
	Logger         *slog.Logger
	AuthMiddleware fiber.Handler
}

// Register wires all routes onto the Fiber app.
func Register(app *fiber.App, deps Deps) {
	// --- Global middleware ---
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(deps.Logger))
	app.Use(middleware.Recover(deps.Logger))
	app.Use(middleware.CORS(deps.Cfg))

	// --- Swagger UI (public, outside /api/v1) ---
	// Serves the UI at /swagger and the spec at openapi.yaml
	openapiPath := resolveOpenAPIPath()
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath: "/",
		Path:     "swagger",
		FilePath: openapiPath,
		Title:    "Portfolio API documentation",
	}))

	api := app.Group("/api/v1")

	// --- Health (no auth, no cache) ---
	api.Get("/health", deps.Health.Health)

	// --- Public read endpoints (cached) ---
	api.Get("/experiences", deps.Experience.List)
	api.Get("/experiences/:id", deps.Experience.Get)

	api.Get("/projects", deps.Project.List)
	api.Get("/projects/:slug", deps.Project.GetBySlug)

	api.Get("/blog/posts", deps.Post.List)
	api.Get("/blog/posts/:slug", deps.Post.GetBySlug)

	api.Get("/tags", deps.Stats.Tags)
	api.Get("/stats/summary", deps.Stats.Summary)

	// --- Profile (public singleton) ---
	api.Get("/profile", deps.Profile.Get)

	// --- Assets (public redirect) ---
	api.Get("/assets/*", deps.Asset.Redirect)

	// --- Auth (rate-limited) ---
	auth := api.Group("/auth")
	auth.Post("/login", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	}), deps.Auth.Login)
	auth.Post("/refresh", deps.Auth.Refresh)

	// --- Admin (auth required) ---
	admin := api.Group("/admin", deps.AuthMiddleware, middleware.RequireAdmin())

	// Experiences CRUD
	admin.Get("/experiences", deps.Experience.List)
	admin.Get("/experiences/:id", deps.Experience.Get)
	admin.Post("/experiences", deps.Experience.Create)
	admin.Put("/experiences/:id", deps.Experience.Update)
	admin.Patch("/experiences/:id", deps.Experience.Update)
	admin.Delete("/experiences/:id", deps.Experience.Delete)

	// Projects CRUD
	admin.Get("/projects", deps.Project.AdminList)
	admin.Get("/projects/:id", deps.Project.AdminGet)
	admin.Post("/projects", deps.Project.Create)
	admin.Put("/projects/:id", deps.Project.Update)
	admin.Patch("/projects/:id", deps.Project.Update)
	admin.Delete("/projects/:id", deps.Project.Delete)

	// Blog posts CRUD
	admin.Get("/blog/posts", deps.Post.AdminList)
	admin.Get("/blog/posts/:id", deps.Post.AdminGet)
	admin.Post("/blog/posts", deps.Post.Create)
	admin.Put("/blog/posts/:id", deps.Post.Update)
	admin.Patch("/blog/posts/:id", deps.Post.Update)
	admin.Delete("/blog/posts/:id", deps.Post.Delete)
	admin.Patch("/blog/posts/:id/status", deps.Post.SetStatus)

	// Assets
	admin.Post("/assets", deps.Asset.Upload)
	admin.Delete("/assets/*", deps.Asset.Delete)

	// Stats (admin)
	admin.Get("/stats/views", deps.Stats.ViewsTimeSeries)
	admin.Get("/stats/top-posts", deps.Stats.TopPosts)

	// Profile (admin singleton — upsert)
	admin.Get("/profile", deps.Profile.AdminGet)
	admin.Put("/profile", deps.Profile.Update)
}

func resolveOpenAPIPath() string {
	candidates := []string{
		"./docs/openapi.yaml",
		"../../docs/openapi.yaml",
		"../docs/openapi.yaml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "./docs/openapi.yaml"
}
