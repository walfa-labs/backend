# Portfolio Backend

Go (Fiber v3) backend API for a personal portfolio with dynamic content — experience, projects, and blog posts. Built with a clean/hexagonal architecture (Ports & Adapters), Sonic JSON, gookit/slog logging, and Oracle Cloud polyglot persistence: Autonomous Transaction Processing (ATP) for operational data, Autonomous Data Warehouse (ADW) for analytics, and OCI Object Storage for assets.

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | [Fiber v3](https://github.com/gofiber/fiber) |
| JSON | [Sonic](https://github.com/bytedance/sonic) (JIT + SIMD) |
| Logging | [gookit/slog](https://github.com/gookit/slog) (rotating files + colored console) |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Database (OLTP) | Oracle ATP via [godror](https://github.com/godror/godror) (database/sql) |
| Database (analytics) | Oracle ADW via [godror](https://github.com/godror/godror) (database/sql) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) (oracle driver) |
| Auth | [golang-jwt/v5](https://github.com/golang-jwt/jwt) |
| Object storage | [oci-go-sdk](https://github.com/oracle/oci-go-sdk) (OCI Object Storage) |
| Config | [caarlos0/env](https://github.com/caarlos0/env) |
| API docs | [contrib/swaggerui](https://github.com/gofiber/contrib) (Swagger UI + OpenAPI 3) |

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
    repository/oracle/
      atp/               → ATP (OLTP) repositories (driven adapter)
      adw/               → ADW analytics store (driven adapter)
      objectstorage/     → OCI Object Storage asset store (driven adapter)
  router/                → route registration
  platform/              → server factory, logger, db pool, json, validator
migrations/
  atp/                   → golang-migrate sequence for the ATP database
  adw/                   → golang-migrate sequence for the ADW database
  seed.sql               → demo data for ATP (applied manually)
```

## Quick Start

Prerequisites: Go 1.26+, [Task](https://taskfile.dev), and **already-running Oracle Cloud infrastructure** — this project does not deploy anything itself. You need (all available in the OCI Always Free tier):

- an **Autonomous Transaction Processing** database and an **Autonomous Data Warehouse** database, with their wallet zip(s) downloaded and unzipped (point `TNS_ADMIN` at the wallet directory),
- an **OCI Object Storage** bucket that already exists, plus an OCI API key pair (unencrypted PEM) for a user with access to it,
- **Oracle Instant Client** installed (godror needs it at runtime; builds work without it),
- a **C compiler for CGO** (`gcc`/`clang`; `zig cc` works) — godror requires `CGO_ENABLED=1`,
- **SQLcl** (`sql`) and the **golang-migrate** CLI for the DB tasks (`task tools` installs the latter, built with the `oracle` tag).

All connection details live in `.env`:

```bash
# Copy the env template, then edit it to point at your infra:
# ATP_DSN, ADW_DSN, TNS_ADMIN, MIGRATE_ATP_URL/MIGRATE_ADW_URL,
# OCI_TENANCY_OCID/USER_OCID/FINGERPRINT/REGION/PRIVATE_KEY_PATH/BUCKET,
# JWT_SECRET, ADMIN_PASSWORD_HASH, ... (see .env.example for everything)
cp .env.example .env

# Install dev tools used below (air + golang-migrate CLI)
task tools

# Apply database migrations to ATP and ADW, then optionally load demo data
task migrate-up
task seed

# Create an admin user — login reads the admin_users table, not
# ADMIN_PASSWORD_HASH; generate your own bcrypt hash for anything real.
sql "$ATP_DSN" <<< \
  "INSERT INTO admin_users (admin_user_id, username, password_hash) VALUES ('<uuid>', 'admin', '<bcrypt-hash>');"

# Start the server on :8080 (Task loads .env automatically)
task run
```

Without Task, the equivalent is to export the variables yourself:
`set -a; . ./.env; set +a && go run ./cmd/api`

Note: there is intentionally no `docker-compose.yml` anymore — the datastores are managed Oracle Cloud services, so there is nothing to run locally in containers.

## API Endpoints

Interactive docs: **Swagger UI at `/swagger`** (public, no auth). The OpenAPI 3 spec
lives at `docs/openapi.yaml` and is served at `/docs/openapi.yaml` — it is read from
disk at startup, so run the binary from the repo root.

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
task migrate-up   # apply migrations to ATP + ADW (MIGRATE_*_URL from .env)
task migrate-down # roll back migrations
task seed         # load demo data into ATP
```
