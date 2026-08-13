# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

Portfolio Backend — a Go (Fiber v3) REST API powering a personal portfolio site with
dynamic content: experiences, projects, blog posts, tags, stats, a singleton profile,
and image assets. Single admin user authenticates via JWT; public endpoints are
read-only and cache-friendly.

Persistence is **polyglot on Oracle Cloud** (all Always Free eligible):
operational OLTP lives in Autonomous Transaction Processing (ATP), view analytics in
Autonomous Data Warehouse (ADW), and asset blobs in OCI Object Storage.

- Module: `github.com/walfa-labs/backend`
- Go version: 1.26.5 (toolchain pinned in `go.mod`)
- Production deploy: single binary built with `task build` (output in `bin/`)

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | `github.com/gofiber/fiber/v3` (v3.4.0, fasthttp-based) |
| JSON | `github.com/bytedance/sonic` (wired via `platform.FiberConfig`) |
| Logging | `github.com/gookit/slog` (+ `gookit/rotatefile` in production) |
| Validation | `github.com/go-playground/validator/v10` (as Fiber `StructValidator`) |
| Database (OLTP) | Oracle ATP via `github.com/godror/godror` (`database/sql`) |
| Database (analytics) | Oracle ADW via `github.com/godror/godror` (`database/sql`) |
| Migrations | golang-migrate CLI, `oracle` driver (plain SQL files in `migrations/`) |
| Auth | `github.com/golang-jwt/jwt/v5` (HS256) + `golang.org/x/crypto/bcrypt` |
| Object storage | `github.com/oracle/oci-go-sdk/v65` (OCI Object Storage) |
| Config | `github.com/caarlos0/env/v11` (environment variables only) |
| API docs | `github.com/gofiber/contrib/v3/swaggerui` (Swagger UI serving `docs/openapi.yaml`) |

## Architecture

Clean/hexagonal architecture (Ports & Adapters). Dependency direction:
`adapter` → `service` → `port` → `domain`. `domain` imports nothing internal.

```
cmd/api/main.go            → entrypoint: manual DI wiring, graceful shutdown
internal/
  config/                  → env loading (config.Config, caarlos0/env struct tags)
  domain/                  → entities, enums, sentinel errors; zero dependencies
  port/                    → interfaces: repository.go (driven ports: *Repo,
                             AssetStore, AnalyticsStore), service.go
                             (use-case ports + input DTOs like ExperienceInput)
  service/                 → application layer; implements port.*Service, holds
                             business rules (input validation, status transitions,
                             stats composition across ATP + ADW)
  adapter/
    handler/               → driving adapter: Fiber handlers, DTO mapping (dto.go),
                             response envelopes (response.go)
    middleware/            → requestid, logger, recover, cors, auth (JWT), error
    repository/oracle/
      atp/                 → driven adapter: godror SQL implementations of
                             port.*Repo against ATP (OLTP)
      adw/                 → driven adapter: port.AnalyticsStore against ADW
      objectstorage/       → driven adapter: port.AssetStore against OCI Object
                             Storage
  router/                  → route registration (router.Register(app, router.Deps{...}))
  platform/                → infrastructure factories: server (Fiber config), Oracle
                             db pool (NewOracleDB), logger, json (Sonic), validator
migrations/
  atp/                     → golang-migrate sequence for the ATP database
  adw/                     → golang-migrate sequence for the ADW database
  seed.sql                 → demo data for ATP, applied manually
docs/                      → OpenAPI 3 spec (openapi.yaml); served by swaggerui
```

Request flow: `router` → `middleware` chain → `handler` (bind + validate request DTO,
map to `port.*Input`) → `service` (business rules) → `port.*Repo` interface →
`atp`/`adw`/`objectstorage` adapter → response mapped back through `dto.go` into the
success envelope.

Key cross-cutting mechanics:

- **DI is manual** in `cmd/api/main.go`: two `*sql.DB` pools (`platform.NewOracleDB`
  for ATP and ADW) → repos → services → handlers → `router.Deps`. To add a feature,
  wire a new repo/service/handler there and extend `router.Deps`.
- **Compile-time port assertions**: each adapter package has `assertions.go` with
  `var _ port.X = (*Y)(nil)` guards. Keep these updated when adding repos.
- **Central error handling**: handlers return errors; `middleware.ErrorHandler`
  (installed as `fiber.Config.ErrorHandler`) maps domain errors to the error envelope.
- **Validation**: request structs use `validate:` tags; `c.Bind().Body(&req)` triggers
  the `platform.StructValidator`, which returns `*domain.ValidationError` → 400 with
  per-field details.
- **View tracking is dual-write** (decision recorded for the Oracle migration):
  `PostService.GetPublishedBySlug` best-effort (1) increments the
  `blog_posts.view_count` counter in ATP — kept so public post responses keep their
  `viewCount` field — and (2) records one `fact_post_views` event in ADW via
  `port.AnalyticsStore`, which is the analytical source of truth for
  `/admin/stats/views`, `/admin/stats/top-posts`, and `Summary.TotalPostViews`.
  Both stores are written on every public post read; neither failure blocks the read.

## API Shape

Base path `/api/v1`. Responses use envelopes (see `internal/adapter/handler/response.go`
and `internal/adapter/middleware/error.go`):

- Success: `{"data": ...}` or `{"data": [...], "meta": {page, perPage, total}}`
- Error: `{"error": {"code": "NOT_FOUND|CONFLICT|VALIDATION_FAILED|UNAUTHORIZED|FORBIDDEN|INTERNAL_ERROR", "message": "...", "details"?: [...], "request_id": "..."}}`

Route groups (see `internal/router/router.go`):

- Public reads: `/experiences`, `/projects`, `/blog/posts`, `/tags`, `/stats/summary`,
  `/profile`, `/assets/*` (302 redirect to an OCI pre-authenticated request URL),
  `/health` (pings both ATP and ADW)
- Auth: `POST /auth/login` (rate-limited: 5 req/min per IP via Fiber limiter),
  `POST /auth/refresh` (reads `refresh_token` httpOnly cookie first, body fallback)
- Admin (`/admin/*`, JWT required): full CRUD for experiences/projects/blog posts,
  `PATCH /admin/blog/posts/:id/status`, asset upload (`POST /admin/assets`, multipart,
  10 MB max, images only) and delete, stats (`/admin/stats/views`,
  `/admin/stats/top-posts` — both served from ADW analytics events), profile upsert
  (`PUT /admin/profile`)
- Docs (outside `/api/v1`, public): Swagger UI at `/swagger`, OpenAPI spec at
  `/docs/openapi.yaml` (wired via `swaggerui.New` in `router.Register`; the spec file
  is read from `./docs/openapi.yaml` at startup — run from the repo root)

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
task migrate-up / task migrate-down / task seed   # databases (MIGRATE_*_URL / ATP_DSN from .env)
task tools           # install air + golang-migrate CLI (oracle build tag)
```

Raw equivalents (`go run ./cmd/api`, `go build ./...`, `go test ./...`, `go vet ./...`,
`go mod tidy`) work as before; note the Taskfile auto-loads `.env` via go-task's
`dotenv`, which the plain `go` commands do not.

**CGO is required** — godror compiles ODPI-C at build time. If `go build` fails with
CGo errors, ensure `CGO_ENABLED=1` and a working C compiler (`CC=gcc`, `CC=clang`, or
`CC="zig cc"`; the machine this migration was done on uses `go env -w CGO_ENABLED=1
CC="zig cc"` because it has no gcc). Oracle Instant Client is needed at **runtime**
(not for builds).

## Running Locally

The process does **not** auto-load `.env` — either run via `task run` (the
Taskfile loads `.env` automatically) or export variables first:

```bash
cp .env.example .env
set -a; . ./.env; set +a        # or: export $(grep -v '^#' .env | xargs)
go run ./cmd/api
```

Required infrastructure — assumed already running; point `.env` at it
(see README "Quick Start"). There is no `docker-compose.yml`: the datastores are
managed Oracle Cloud services.

1. ATP + ADW reachable via `ATP_DSN` / `ADW_DSN` godror connect strings
   (`user/password@tns_alias`), with `TNS_ADMIN` pointing at the unzipped wallet dir
2. OCI Object Storage bucket created + API key PEM for a user with access
3. Oracle Instant Client installed (runtime), CGO toolchain (build)
4. Migrations applied to **both** databases: `task migrate-up`
   (`migrate -path migrations/atp -database "$MIGRATE_ATP_URL" up` and the same for
   `migrations/adw` with `MIGRATE_ADW_URL`)
5. An `admin_users` row with a bcrypt `password_hash` (login reads the DB table,
   not `ADMIN_PASSWORD_HASH`); `migrations/seed.sql` holds demo content data
   (load with `task seed`, needs SQLcl)

## Configuration

All config is environment variables parsed by `internal/config/config.go`.
`.env.example` documents every variable. Required (no default): `ATP_DSN`, `ADW_DSN`,
`JWT_SECRET`, `ADMIN_PASSWORD_HASH`, and the `OCI_*` group (`TENANCY_OCID`,
`USER_OCID`, `FINGERPRINT`, `REGION`, `PRIVATE_KEY_PATH`, `BUCKET`; `OCI_NAMESPACE`
optional — resolved via the GetNamespace API when empty). `TNS_ADMIN` and
`MIGRATE_ATP_URL` / `MIGRATE_ADW_URL` are consumed by the Oracle client /
golang-migrate respectively, not by `config.go`. Notable: `APP_ENV`
(`development`|`production` switches logger behavior), `APP_PORT` (default `:8080`),
`JWT_ACCESS_TTL`/`JWT_REFRESH_TTL`, `CORS_ALLOWED_ORIGINS` (comma-separated).

## Code Style Guidelines

- Standard Go: `gofmt`-clean, `go vet`-clean. Package comments on exported types/funcs
  follow the existing doc-comment style (concise, references to "design doc §x.y" appear
  in a few places).
- **Layers**: keep the hexagonal boundaries — `domain` never imports other internal
  packages; services depend on `port` interfaces, not concrete repos; SQL lives only in
  `adapter/repository/oracle/atp` (+ `.../adw`); HTTP concerns only in
  `adapter/handler` + `middleware`.
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
- **Repositories (atp/adw)**: raw SQL with godror positional placeholders (`:1, :2`)
  via `database/sql` (`*sql.DB`, `*sql.Tx`); `sql.ErrNoRows` → `domain.ErrNotFound`;
  multi-table writes (post+tags, experience+highlights, project+links, ADW
  dim+fact upsert) run in a `*sql.Tx`; `RowsAffected() == 0` on UPDATE/DELETE →
  `domain.ErrNotFound`. Oracle idioms (see `atp/helpers.go`): booleans are NUMBER(1)
  (`b2i`/`i2b`), nullable text scans via `sql.NullString` (`nullStr`), nullable
  timestamps via `sql.NullTime` (`nullTime`), CLOB binds via `clob()` (plain string
  binds cap at 32767 bytes), UUIDs bind as `.String()` and scan directly into
  `uuid.UUID`, pagination is `OFFSET :n ROWS FETCH NEXT :m ROWS ONLY`, upserts are
  `MERGE INTO`, and `experiences."current"` must stay double-quoted (reserved word).
- **Naming**: DB tables are plural snake_case (`blog_posts`, `experience_highlights`)
  with `{table}_id` primary keys (VARCHAR2(36 CHAR), service-generated UUIDs);
  Go entities singular (`domain.BlogPost`); DTO JSON fields camelCase.
- **IDs/timestamps**: UUIDs generated in the service layer (`uuid.New()`); child-row
  IDs (highlights, links, ADW fact rows) generated in the repo layer. `created_at` /
  `updated_at` default to `CURRENT_TIMESTAMP` in DDL; repos re-SELECT them by PK after
  insert (no RETURNING) and set `updated_at = CURRENT_TIMESTAMP` in UPDATEs.

## Testing Instructions

- There is currently **no test suite** (`go test ./...` compiles but runs nothing) and
  no CI configuration in the repo. When adding tests, follow standard Go conventions:
  `*_test.go` beside the code, table-driven tests, and fake implementations of the
  `port.*` interfaces for service-layer tests (ports exist precisely to enable this).
- Minimum verification before finishing a change: `go build ./...` and `go vet ./...`.

## Database & Migrations

- Migrations are golang-migrate SQL files in **two sequences** (two databases):
  `migrations/atp/NNNN_name.{up,down}.sql` (OLTP schema: 0001_init, 0002_profile) and
  `migrations/adw/NNNN_name.{up,down}.sql` (analytics star schema: 0001_analytics with
  `dim_posts` + `fact_post_views`). `task migrate-up`/`migrate-down` run both.
  `migrations/seed.sql` is demo data for ATP, applied manually via SQLcl — not part
  of either migrate sequence.
- Schema notes (Oracle translation of the old Postgres schema): enums are
  `VARCHAR2(20)` + CHECK (no TYPE objects); booleans `NUMBER(1)` + CHECK;
  `projects.tech_stack` is a JSON array string in `VARCHAR2(4000)` with
  `CHECK (tech_stack IS JSON)`; long markdown is CLOB; the Postgres partial index on
  `blog_posts.published_at` became a plain descending index; `profiles` stays a
  singleton (`CHECK (profile_id = 1)`), with CLOB columns `NOT NULL DEFAULT
  EMPTY_CLOB()` because Oracle stores `''` as NULL.
- Never edit an applied migration — add a new numbered pair instead (in the sequence
  of the database it belongs to).

## Security Considerations

- JWT is HS256 with `JWT_SECRET`; access token (default 15m) returned in the body,
  refresh token (default 7d) also set as an httpOnly, Secure, SameSite=Strict cookie
  scoped to `Path=/api/v1/auth`. Login is rate-limited (5/min/IP).
- Admin passwords are bcrypt hashes; login compares against the `admin_users` table.
- `middleware.Auth` is non-blocking (populates claims when a valid Bearer token exists);
  admin routes are protected by the `/admin` group's middleware chain. Keep new admin
  routes under that group.
- Asset uploads are restricted to image content types (jpeg/png/gif/webp/avif/svg) and
  10 MB; public asset access goes through DB lookup + an OCI pre-authenticated
  request (PAR) URL with 1 h expiry (PARs are named resources that accumulate
  server-side until expiry — unlike stateless S3 presignatures).
- All SQL is parameterized — keep it that way. CORS origins are an explicit allowlist.
- Never commit `.env`, wallet files, or the OCI API key PEM (`.gitignore` covers
  `.env`; keep `wallet/` and `*.pem` out of commits too); the error envelope
  intentionally hides internal error causes from clients.

## Gotchas

- Fiber **v3** API (`fiber.Ctx` is an interface, `c.Bind().Body(...)`,
  `c.Redirect().To(...)`) — do not copy Fiber v2 idioms.
- godror needs **CGO at build time** (a C compiler: gcc/clang/`zig cc`) and **Oracle
  Instant Client at runtime**. A missing C compiler shows up as
  `build constraints exclude all Go files in .../godror` or CGo errors.
- Oracle stores `''` as **NULL** — never rely on empty-string semantics; scan nullable
  text with `sql.NullString`. Conversely, binding a Go `""` to a VARCHAR2 column
  stores NULL (harmless with the NullString read pattern).
- godror fetches CLOBs as strings **by default**; for binds use the `clob()` helper
  (`godror.Lob{Reader: ..., IsClob: true}`) so bodies > 32767 bytes work.
- `"current"` (experiences table) is an Oracle reserved word — always double-quoted,
  in DDL, seed data, and Go SQL.
- The OCI Object Storage "presign" is a **pre-authenticated request (PAR)**, created
  per call with 1 h expiry; PARs pile up in the bucket until they expire.
- `TNS_ADMIN` is read by the Oracle client libraries, not by `config.go` — set it in
  the environment (or `.env` via the Taskfile) pointing at the wallet directory.
- Sonic is the JSON codec; it works CGO-free on amd64/arm64 (the CGO requirement
  comes from godror, not Sonic).
- Production logging writes rotating files under `logs/` (created at runtime).
