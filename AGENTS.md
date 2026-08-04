# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

Portfolio Backend — a Go (Fiber v3) REST API powering a personal portfolio site with
dynamic content: experiences, projects, blog posts, tags, stats, a singleton profile,
and image assets. Single admin user authenticates via JWT; public endpoints are
read-only and cache-friendly.

- Module: `github.com/walfa-labs/backend`
- Go version: 1.26.5 (toolchain pinned in `go.mod`)
- Production deploy: single static binary built with `task build` (output in `bin/`)

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | `github.com/gofiber/fiber/v3` (v3.4.0, fasthttp-based) |
| JSON | `github.com/bytedance/sonic` (wired via `platform.FiberConfig`) |
| Logging | `github.com/gookit/slog` (+ `gookit/rotatefile` in production) |
| Validation | `github.com/go-playground/validator/v10` (as Fiber `StructValidator`) |
| Database | PostgreSQL 16+ via `github.com/jackc/pgx/v5` (`pgxpool`) |
| Migrations | golang-migrate CLI (plain SQL files in `migrations/`) |
| Auth | `github.com/golang-jwt/jwt/v5` (HS256) + `golang.org/x/crypto/bcrypt` |
| Object storage | `github.com/minio/minio-go/v7` (S3-compatible: MinIO, R2, S3) |
| Config | `github.com/caarlos0/env/v11` (environment variables only) |

## Architecture

Clean/hexagonal architecture (Ports & Adapters). Dependency direction:
`adapter` → `service` → `port` → `domain`. `domain` imports nothing internal.

```
cmd/api/main.go            → entrypoint: manual DI wiring, graceful shutdown
internal/
  config/                  → env loading (config.Config, caarlos0/env struct tags)
  domain/                  → entities, enums, sentinel errors; zero dependencies
  port/                    → interfaces: repository.go (driven ports), service.go
                             (use-case ports + input DTOs like ExperienceInput)
  service/                 → application layer; implements port.*Service, holds
                             business rules (input validation, status transitions)
  adapter/
    handler/               → driving adapter: Fiber handlers, DTO mapping (dto.go),
                             response envelopes (response.go)
    middleware/            → requestid, logger, recover, cors, auth (JWT), error
    repository/postgres/   → driven adapter: pgx SQL implementations of port.*Repo
    repository/s3/         → driven adapter: minio-go implementation of port.AssetStore
  router/                  → route registration (router.Register(app, router.Deps{...}))
  platform/                → infrastructure factories: server (Fiber config), db pool,
                             logger, json (Sonic), validator
migrations/                → NNNN_name.up.sql / NNNN_name.down.sql + seed.sql
```

Request flow: `router` → `middleware` chain → `handler` (bind + validate request DTO,
map to `port.*Input`) → `service` (business rules) → `port.*Repo` interface →
`postgres`/`s3` adapter → response mapped back through `dto.go` into the success envelope.

Key cross-cutting mechanics:

- **DI is manual** in `cmd/api/main.go`: repos → services → handlers → `router.Deps`.
  To add a feature, wire a new repo/service/handler there and extend `router.Deps`.
- **Compile-time port assertions**: each adapter package has `assertions.go` with
  `var _ port.X = (*Y)(nil)` guards. Keep these updated when adding repos.
- **Central error handling**: handlers return errors; `middleware.ErrorHandler`
  (installed as `fiber.Config.ErrorHandler`) maps domain errors to the error envelope.
- **Validation**: request structs use `validate:` tags; `c.Bind().Body(&req)` triggers
  the `platform.StructValidator`, which returns `*domain.ValidationError` → 400 with
  per-field details.

## API Shape

Base path `/api/v1`. Responses use envelopes (see `internal/adapter/handler/response.go`
and `internal/adapter/middleware/error.go`):

- Success: `{"data": ...}` or `{"data": [...], "meta": {page, perPage, total}}`
- Error: `{"error": {"code": "NOT_FOUND|CONFLICT|VALIDATION_FAILED|UNAUTHORIZED|FORBIDDEN|INTERNAL_ERROR", "message": "...", "details"?: [...], "request_id": "..."}}`

Route groups (see `internal/router/router.go`):

- Public reads: `/experiences`, `/projects`, `/blog/posts`, `/tags`, `/stats/summary`,
  `/profile`, `/assets/*` (302 redirect to presigned S3 URL), `/health`
- Auth: `POST /auth/login` (rate-limited: 5 req/min per IP via Fiber limiter),
  `POST /auth/refresh` (reads `refresh_token` httpOnly cookie first, body fallback)
- Admin (`/admin/*`, JWT required): full CRUD for experiences/projects/blog posts,
  `PATCH /admin/blog/posts/:id/status`, asset upload (`POST /admin/assets`, multipart,
  10 MB max, images only) and delete, stats (`/admin/stats/views`, `/admin/stats/top-posts`),
  profile upsert (`PUT /admin/profile`)

JSON field names are camelCase (e.g. `coverImageUrl`, `sortOrder`, `experienceType`).
Dates in requests are `YYYY-MM-DD`; timestamps in responses are RFC3339.

## Build and Test Commands

All common commands are Taskfile tasks (`task --list` to see them):

```bash
task run             # start the server (default :8080, loads .env)
task build           # build the binary to bin/ (-ldflags="-s -w")
task test            # run tests (currently no *_test.go files exist)
task vet             # static analysis
task tidy            # sync dependencies
task dev             # hot reload via air (see .air.toml)
task migrate-up / task migrate-down / task seed   # database (DATABASE_URL from .env)
task tools           # install air + golang-migrate CLI
```

Raw equivalents (`go run ./cmd/api`, `go build ./...`, `go test ./...`, `go vet ./...`,
`go mod tidy`) work as before; note the Taskfile auto-loads `.env` via go-task's
`dotenv`, which the plain `go` commands do not.

## Running Locally

The process does **not** auto-load `.env` — either run via `task run` (the
Taskfile loads `.env` automatically) or export variables first:

```bash
cp .env.example .env
set -a; . ./.env; set +a        # or: export $(grep -v '^#' .env | xargs)
go run ./cmd/api
```

Required infrastructure — assumed already running; point `.env` at it
(see README "Quick Start"):

1. PostgreSQL 16+ reachable via `DATABASE_URL`
2. S3-compatible store (MinIO locally) with the bucket created
3. Migrations applied: `migrate -path migrations -database "$DATABASE_URL" up`
4. An `admin_user` row with a bcrypt `password_hash` (login reads the DB table,
   not `ADMIN_PASSWORD_HASH`); `migrations/seed.sql` holds demo content data

## Configuration

All config is environment variables parsed by `internal/config/config.go`.
`.env.example` documents every variable. Required (no default): `DATABASE_URL`,
`JWT_SECRET`, `ADMIN_PASSWORD_HASH`. Notable: `APP_ENV` (`development`|`production`
switches logger behavior), `APP_PORT` (default `:8080`), `JWT_ACCESS_TTL`/`JWT_REFRESH_TTL`,
`OBJECT_STORAGE_*` (endpoint may carry an `http://`/`https://` scheme; bare host means TLS),
`CORS_ALLOWED_ORIGINS` (comma-separated).

## Code Style Guidelines

- Standard Go: `gofmt`-clean, `go vet`-clean. Package comments on exported types/funcs
  follow the existing doc-comment style (concise, references to "design doc §x.y" appear
  in a few places).
- **Layers**: keep the hexagonal boundaries — `domain` never imports other internal
  packages; services depend on `port` interfaces, not concrete repos; SQL lives only in
  `adapter/repository/postgres`; HTTP concerns only in `adapter/handler` + `middleware`.
- **Errors**: return `domain` sentinel errors (`ErrNotFound`, `ErrConflict`,
  `ErrValidation`, `ErrUnauthorized`, `ErrForbidden`) or `domain.NewValidationError(
  "field", "issue", ...)`. Wrap infra errors with `fmt.Errorf("context: %w", err)`.
  Never write error responses directly in handlers — return the error and let the
  central `ErrorHandler` render it.
- **Constructors**: `NewX(dep1, dep2) *X` per file (`NewPostService`, `NewPostRepo`,
  `NewPostHandler`, ...). Handlers/services/repos are concrete structs; interfaces are
  consumed (ports), not re-declared locally.
- **Handlers**: parse params (`uuid.Parse(c.Params("id"))` → 400 on failure),
  `c.Bind().Body(&req)` (invalid JSON → `domain.NewValidationError("body", "invalid JSON")`),
  map request → `port.*Input` via a `toInput()` method on the request DTO, call the
  service, map the result via a `toXResponse()` function in `dto.go`.
- **Responses**: use the helpers in `handler/response.go` — `OK`, `OKWithMeta`,
  `Created` (201 + `Location`), `NoContent` (204). Public GETs set
  `PublicCacheHeaders(c, etag)`; mutations and admin endpoints set `NoStoreHeaders(c)`.
- **Repositories**: raw SQL with `$n` placeholders via `pgxpool`; `pgx.ErrNoRows` →
  `domain.ErrNotFound`; multi-table writes (post+tags, experience+highlights,
  project+links) run in a `pgx.Tx`; nullable columns are read with `COALESCE(col, '')`;
  `RowsAffected() == 0` on UPDATE/DELETE → `domain.ErrNotFound`.
- **Naming**: DB tables are singular snake_case (`blog_post`, `experience_highlight`);
  Go entities singular (`domain.BlogPost`); DTO JSON fields camelCase.
- **IDs/timestamps**: UUIDs generated in the service layer (`uuid.New()`); DB defaults
  also exist (`gen_random_uuid()`). `updated_at` maintained in SQL (`now()`).

## Testing Instructions

- There is currently **no test suite** (`go test ./...` compiles but runs nothing) and
  no CI configuration in the repo. When adding tests, follow standard Go conventions:
  `*_test.go` beside the code, table-driven tests, and fake implementations of the
  `port.*` interfaces for service-layer tests (ports exist precisely to enable this).
- Minimum verification before finishing a change: `go build ./...` and `go vet ./...`.

## Database & Migrations

- Migrations are golang-migrate SQL files: `migrations/NNNN_name.up.sql` /
  `NNNN_name.down.sql` (currently `0001_init`, `0002_profile`). `migrations/seed.sql`
  is demo data, applied manually — it is not part of the migrate sequence.
- Schema highlights: enum types `experience_type` (`work|education`) and
  `content_status` (`draft|published`); `post_tag` join table; `profile` is a
  singleton enforced by `CHECK (id = 1)`; cascades on child tables.
- Never edit an applied migration — add a new numbered pair instead.

## Security Considerations

- JWT is HS256 with `JWT_SECRET`; access token (default 15m) returned in the body,
  refresh token (default 7d) also set as an httpOnly, Secure, SameSite=Strict cookie
  scoped to `Path=/api/v1/auth`. Login is rate-limited (5/min/IP).
- Admin passwords are bcrypt hashes; login compares against the `admin_user` table.
- `middleware.Auth` is non-blocking (populates claims when a valid Bearer token exists);
  admin routes are protected by the `/admin` group's middleware chain. Keep new admin
  routes under that group.
- Asset uploads are restricted to image content types (jpeg/png/gif/webp/avif/svg) and
  10 MB; public asset access goes through DB lookup + presigned URL (1h expiry).
- All SQL is parameterized — keep it that way. CORS origins are an explicit allowlist.
- Never commit `.env` or real secrets (`.gitignore` already covers it); the error
  envelope intentionally hides internal error causes from clients.

## Gotchas

- Fiber **v3** API (`fiber.Ctx` is an interface, `c.Bind().Body(...)`,
  `c.Redirect().To(...)`) — do not copy Fiber v2 idioms.
- Sonic is the JSON codec; it works CGO-free on amd64/arm64.
- The S3 endpoint config accepts a scheme prefix; minio-go must receive a bare host
  (handled by `splitEndpointScheme` in `adapter/repository/s3/asset_store.go`).
- Production logging writes rotating files under `logs/` (created at runtime).
