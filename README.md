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

Prerequisites: Go 1.26+, [Task](https://taskfile.dev), and **already-running infrastructure** — this project does not deploy anything itself. You need:

- a PostgreSQL 16+ database you can reach (local install, home server, managed — anything),
- an S3-compatible bucket (MinIO, R2, S3) that already exists,
- `psql` and the `golang-migrate` CLI for the DB tasks (`task tools` installs the latter).

All connection details live in `.env`:

```bash
# Copy the env template, then edit it to point at your infra:
# DATABASE_URL, OBJECT_STORAGE_ENDPOINT/BUCKET/ACCESS_KEY/SECRET_KEY,
# JWT_SECRET, ADMIN_PASSWORD_HASH, ... (see .env.example for everything)
cp .env.example .env

# Install dev tools used below (air + golang-migrate CLI)
task tools

# Apply database migrations, then optionally load demo data
task migrate-up
task seed

# Create an admin user — login reads the admin_user table, not
# ADMIN_PASSWORD_HASH; generate your own bcrypt hash for anything real.
psql "$DATABASE_URL" -c \
  "INSERT INTO admin_user (username, password_hash) VALUES ('admin', '<bcrypt-hash>');"

# Start the server on :8080 (Task loads .env automatically)
task run
```

Without Task, the equivalent is to export the variables yourself:
`set -a; . ./.env; set +a && go run ./cmd/api`

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
GET /api/v1/assets/*            → redirect to signed URL (key may contain slashes, e.g. images/<uuid>.png)
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
DELETE /api/v1/admin/assets/*                   → key may contain slashes

GET    /api/v1/admin/stats/views?from=&to=&bucket=
GET    /api/v1/admin/stats/top-posts?limit=
```

## Configuration

See `.env.example` for all environment variables.

## Development

Everything goes through the Taskfile (`task --list` to see all tasks):

```bash
task run          # start the server (loads .env)
task dev          # hot reload via air
task build        # build the binary to bin/
task test         # go test ./...
task vet          # go vet ./...
task tidy         # go mod tidy
task migrate-up   # apply migrations (DATABASE_URL from .env)
task migrate-down # roll back migrations
task seed         # load demo data
```
