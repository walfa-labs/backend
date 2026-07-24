# Portfolio Backend

Go (Fiber v3) backend API for a personal portfolio with dynamic content — experience, projects, and blog posts. Built with a clean/hexagonal architecture (Ports & Adapters), Sonic JSON, gookit/slog logging, PostgreSQL, and S3-compatible object storage.

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | [Fiber v3](https://github.com/gofiber/fiber) |
| JSON | [Sonic](https://github.com/bytedance/sonic) (JIT + SIMD) |
| Logging | [gookit/slog](https://github.com/gookit/slog) (rotating files + colored console) |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Database | [pgx/v5](https://github.com/jackc/pgx) + PostgreSQL 16+ |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Auth | [golang-jwt/v5](https://github.com/golang-jwt/jwt) |
| Object storage | [minio-go/v7](https://github.com/minio/minio-go) (S3-compatible) |
| Config | [caarlos0/env](https://github.com/caarlos0/env) |

## Architecture

```
cmd/api/main.go          → entrypoint: wire DI, start server
internal/
  config/                → env loading
  domain/                → core entities + errors (no deps)
  port/                  → interfaces (repository + service contracts)
  service/               → use-case / application layer
  adapter/
    handler/             → HTTP handlers (Fiber driving adapter)
    middleware/          → recover, logger, cors, auth, error, requestid
    repository/postgres/ → Postgres implementations (driven adapter)
    repository/s3/       → S3 asset store (driven adapter)
  router/                → route registration
  platform/              → server factory, logger, db pool, json, validator
migrations/              → golang-migrate SQL files
```

## Quick Start

```bash
# Copy env template
cp .env.example .env

# Run database migrations
migrate -path migrations -database "$DATABASE_URL" up

# Start the server
go run ./cmd/api    # serves :8080
```

## API Endpoints

### Public (read-only, cached)
```
GET /api/v1/experiences
GET /api/v1/experiences/:id
GET /api/v1/projects
GET /api/v1/projects/:slug
GET /api/v1/blog/posts
GET /api/v1/blog/posts/:slug
GET /api/v1/tags
GET /api/v1/stats/summary
GET /api/v1/assets/*          → redirect to signed URL
GET /api/v1/health
```

### Auth
```
POST /api/v1/auth/login           → { accessToken, refreshToken }
POST /api/v1/auth/refresh
```

### Admin (JWT required)
```
POST   /api/v1/admin/experiences
GET    /api/v1/admin/experiences
GET    /api/v1/admin/experiences/:id
PUT    /api/v1/admin/experiences/:id
PATCH  /api/v1/admin/experiences/:id
DELETE /api/v1/admin/experiences/:id

POST   /api/v1/admin/projects
GET    /api/v1/admin/projects
GET    /api/v1/admin/projects/:id
PUT    /api/v1/admin/projects/:id
PATCH  /api/v1/admin/projects/:id
DELETE /api/v1/admin/projects/:id

POST   /api/v1/admin/blog/posts
GET    /api/v1/admin/blog/posts
GET    /api/v1/admin/blog/posts/:id
PUT    /api/v1/admin/blog/posts/:id
PATCH  /api/v1/admin/blog/posts/:id
DELETE /api/v1/admin/blog/posts/:id
PATCH  /api/v1/admin/blog/posts/:id/status    → { status: "published"|"draft" }

POST   /api/v1/admin/assets                     → multipart upload
DELETE /api/v1/admin/assets/*

GET    /api/v1/admin/stats/views?from=&to=&bucket=
GET    /api/v1/admin/stats/top-posts?limit=
```

## Configuration

See `.env.example` for all environment variables.

## Development

```bash
go mod tidy
go run ./cmd/api
go test ./...
go vet ./...
```
