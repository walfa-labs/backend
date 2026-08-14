# AGENTS.md

Operational guidance, architectural rules, coding standards, and verification checklists for AI coding assistants working in this repository.

---

## 1. Project Overview & Tech Stack

**Portfolio Backend** is a high-performance REST API written in Go with **Fiber v3**, powering dynamic portfolio platforms (experiences, projects, blog posts, tags, singleton profile, stats, and asset management).

Persistence is polyglot on Oracle Cloud (ATP for OLTP, ADW for analytics, OCI Object Storage for assets) with hybrid support for local development using **Oracle Database 23ai Free in Docker** and **local disk storage**.

### Technology Matrix

| Concern | Library / Tool | Version / Notes |
|---|---|---|
| **Language & Toolchain** | Go | `1.26.6` (CGO required for godror / ODPI-C) |
| **HTTP Framework** | `github.com/gofiber/fiber/v3` | `v3.4.0` (fasthttp core; Fiber v3 APIs only) |
| **JSON Serialization** | `github.com/bytedance/sonic` | SIMD/JIT accelerated parser |
| **Logging** | `github.com/gookit/slog` | Rotating files in prod, colored console in dev |
| **Input Validation** | `github.com/go-playground/validator/v10` | Integrated as Fiber `StructValidator` |
| **Database (OLTP)** | Oracle ATP / Oracle 23ai Free | Native `database/sql` via `github.com/godror/godror` |
| **Database (Analytics)** | Oracle ADW / Oracle 23ai Free | Native `database/sql` via `github.com/godror/godror` |
| **Schema Migrations** | `golang-migrate/migrate/v4` | CLI with `oracle` driver tag; dual sequences |
| **Authentication** | `github.com/golang-jwt/jwt/v5` + `bcrypt` | HS256 JWT (15m access) + httpOnly cookie (7d refresh) |
| **Object Storage** | `github.com/oracle/oci-go-sdk/v65` / Local | OCI Object Storage (PAR presigned URLs) or Local Disk |
| **Configuration** | `github.com/caarlos0/env/v11` | Pure environment variable parsing |
| **API Documentation** | `github.com/gofiber/contrib/v3/swaggerui` | Swagger UI serving `docs/openapi.yaml` at `/swagger` |
| **Task Automation** | [Taskfile](https://taskfile.dev) | Modern task runner (`Taskfile.yml`) |

---

## 2. Hexagonal Architecture & Layer Boundaries

The codebase strictly adheres to **Hexagonal / Clean Architecture (Ports & Adapters)**.

```
cmd/api/main.go               → Composition root / Manual DI wiring, graceful shutdown
internal/
  config/                     → Environment configuration loader (caarlos0/env struct tags)
  domain/                     → Core entities, enums, sentinel errors (ZERO internal/external dependencies)
  port/                       → Driven repository interfaces (*Repo, AssetStore, AnalyticsStore)
                                + use-case service interfaces (*Service) & input DTOs
  service/                    → Application business logic, entity transitions, dual-write view tracking
  adapter/
    handler/                  → Driving adapter: Fiber HTTP handlers, DTO binding, response envelopes
    middleware/               → RequestID, Logger, Recover, CORS, Auth (JWT), SecurityHeaders, ErrorHandler
    repository/
      oracle/atp/             → Driven adapter: godror SQL implementation of port.*Repo (OLTP)
      oracle/adw/             → Driven adapter: godror SQL implementation of port.AnalyticsStore (OLAP)
      oracle/objectstorage/   → Driven adapter: port.AssetStore against OCI Object Storage
      localstorage/           → Driven adapter: port.AssetStore against local filesystem
      memory/                 → Driven adapter: in-memory mock store for tests & DAST mode
  platform/                   → Infrastructure factories: Fiber server, Oracle DB pool, logger, validator
  router/                     → Route registration, middleware mounting, Swagger UI, static uploads
migrations/
  atp/                        → golang-migrate sequence for operational OLTP database
  adw/                        → golang-migrate sequence for analytics OLAP database
  seed.sql / seed_ora.sql     → Demo dataset for ATP
docs/                         → OpenAPI 3.0 specification (openapi.yaml)
```

### Layer Dependency Invariants

1. **`internal/domain`**:
   - Contains pure structs, enums, and sentinel errors (`ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrForbidden`, `ErrValidation`, `NewValidationError`).
   - **Rule**: `domain` MUST NEVER import any other `internal/*` package or external HTTP/DB frameworks.
2. **`internal/port`**:
   - Defines interface contracts for driving and driven boundaries (`port.*Repo`, `port.AssetStore`, `port.AnalyticsStore`, `port.*Service`) and service input DTOs (e.g. `port.ExperienceInput`).
3. **`internal/service`**:
   - Implements use cases from `port.*Service`.
   - Depends ONLY on `domain` and `port` interfaces. Never import concrete adapters (`adapter/*`) or database packages.
   - Generates domain entity UUIDs (`uuid.New()`) before persisting.
4. **`internal/adapter/handler`**:
   - Parses HTTP parameters, binds JSON bodies via `c.Bind().Body(&req)`, validates inputs, maps request DTOs to `port.*Input`, executes services, and maps outputs to response DTOs.
   - Handlers MUST return `error` and NEVER construct error responses manually.
5. **`internal/adapter/repository/*`**:
   - Implements `port.*Repo`, `port.AnalyticsStore`, and `port.AssetStore`.
   - All SQL statements live strictly within `adapter/repository/oracle/atp` and `.../adw`.
   - Every repository package must include compile-time type assertions (`var _ port.X = (*Y)(nil)`).

---

## 3. Critical Oracle SQL & Driver Guidelines

The database adapters use `github.com/godror/godror` over `database/sql`. You MUST follow these Oracle-specific conventions:

1. **Positional Placeholders**:
   - Use `:1`, `:2`, `:3`, ... (godror format). Do NOT use `?` (MySQL) or `$1` (Postgres).
2. **Booleans are `NUMBER(1)`**:
   - Oracle has no native BOOLEAN column type in standard SQL table schemas.
   - Use the package helpers `b2i(bool) int` (returns `1` or `0`) and `i2b(int) bool`.
   - NEVER bind a Go `bool` directly to a SQL query.
3. **Empty String is `NULL`**:
   - Oracle treats `''` as `NULL`.
   - All nullable string columns MUST scan into `sql.NullString` and be unwrapped using `nullStr(ns)`.
   - Nullable timestamps MUST scan into `sql.NullTime` and be unwrapped using `nullTime(nt)`.
4. **CLOB Handling**:
   - Godror scans CLOBs into Go `string` automatically.
   - When binding strings > 32767 bytes (e.g., large markdown bodies), you MUST use the `clob()` helper:
     ```go
     func clob(s string) godror.Lob {
         return godror.Lob{Reader: strings.NewReader(s), IsClob: true}
     }
     ```
5. **Quoted Reserved Words**:
   - The column `"current"` in the `experiences` table is an Oracle SQL reserved keyword. It MUST ALWAYS be double-quoted in all DDL, DML, and SQL queries (e.g., `SELECT "current", start_date FROM experiences`).
6. **Pagination Syntax**:
   - Use ANSI standard Oracle pagination:
     ```sql
     OFFSET :1 ROWS FETCH NEXT :2 ROWS ONLY
     ```
7. **Upsert Semantics**:
   - Use Oracle `MERGE INTO`:
     ```sql
     MERGE INTO profiles p
     USING DUAL ON (p.profile_id = :1)
     WHEN MATCHED THEN
       UPDATE SET name = :2, bio = :3, ...
     WHEN NOT MATCHED THEN
       INSERT (profile_id, name, bio, ...) VALUES (:1, :2, :3, ...)
     ```
8. **No RETURNING Clauses**:
   - Do not rely on `RETURNING` clauses.
   - Insert rows with service-generated UUIDs and re-SELECT timestamps by primary key (`created_at`, `updated_at`).
9. **Transactions**:
   - Any multi-table write (e.g., blog post + tags, experience + highlights, project + links) MUST be executed within a `*sql.Tx`.
10. **Error Mapping**:
    - `sql.ErrNoRows` &rarr; return `domain.ErrNotFound`.
    - If `result.RowsAffected() == 0` on `UPDATE` or `DELETE` &rarr; return `domain.ErrNotFound`.

---

## 4. Fiber v3 API & Handler Conventions

This repository uses **Fiber v3** (`github.com/gofiber/fiber/v3`). Do not use Fiber v2 patterns.

### Key Fiber v3 Idioms

- **Context is an Interface**: Use `c fiber.Ctx`, NOT `*fiber.Ctx`.
- **Body Binding**: Use `if err := c.Bind().Body(&req); err != nil { return domain.NewValidationError("body", "invalid request body") }`.
- **Redirects**: Use `return c.Redirect().To(url)`.
- **Status Constants**: Use Fiber v3 status constants (e.g., `fiber.StatusOK`, `fiber.StatusCreated`, `fiber.StatusNoContent`, `fiber.StatusBadRequest`).

### Response Envelopes (`internal/adapter/handler/response.go`)

- **Single Item Success**:
  ```go
  return response.OK(c, toPostResponse(post)) // {"data": {...}}
  ```
- **List Item Success with Metadata**:
  ```go
  return response.OKWithMeta(c, items, response.Meta{Page: page, PerPage: perPage, Total: total})
  ```
- **Creation Success (201 Created)**:
  ```go
  return response.Created(c, "/api/v1/admin/experiences/"+id.String(), toExperienceResponse(exp))
  ```
- **Deletion Success (204 No Content)**:
  ```go
  return response.NoContent(c)
  ```
- **Cache Headers**:
  - Public GET endpoints: Set `response.PublicCacheHeaders(c, maxAgeSeconds)`.
  - Admin & Mutation endpoints: Set `response.NoStoreHeaders(c)`.

### Centralized Error Handling

Handlers MUST return errors directly. The central `middleware.ErrorHandler` intercepts all errors and maps them into the standard error envelope:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found",
    "details": [],
    "request_id": "78f1889e-c9c3-4ff5-becc-163cc76a1b4b"
  }
}
```

Standard domain error mappings:
- `domain.ErrNotFound` &rarr; `404 NOT_FOUND`
- `domain.ErrConflict` &rarr; `409 CONFLICT`
- `domain.ErrUnauthorized` &rarr; `401 UNAUTHORIZED`
- `domain.ErrForbidden` &rarr; `403 FORBIDDEN`
- `*domain.ValidationError` &rarr; `400 VALIDATION_FAILED` (with field-level error details)
- Unhandled / Infra errors &rarr; `500 INTERNAL_ERROR` (internal stack/cause hidden from client)

---

## 5. Dual-Write View Tracking & Analytics

View tracking on public blog posts implements a dual-write mechanism in `PostService.GetPublishedBySlug`:

1. **ATP Counter (OLTP)**: Atomically increments `blog_posts.view_count` so the public post response immediately reflects the updated count.
2. **ADW Event (OLAP)**: Records a dimensional event in `fact_post_views` (ADW) via `port.AnalyticsStore.RecordPostView`.
3. **Fault Tolerance**: Both write operations are non-blocking / best-effort. If either database write fails, the error is logged, and the post is still returned to the reader successfully.

---

## 6. Asset Storage Architecture

The application supports two asset storage drivers switchable via `STORAGE_DRIVER`:

1. **`STORAGE_DRIVER=local` (Default for Local Development)**:
   - Implemented in `internal/adapter/repository/localstorage`.
   - Saves uploaded files to `STORAGE_LOCAL_DIR` (e.g. `./uploads`).
   - Returns a direct URL (e.g. `http://localhost:8080/uploads/images/<uuid>.png`).
   - Fiber statically routes `/uploads/*` to `c.SendFile(...)`.
2. **`STORAGE_DRIVER=oci` (Production on Oracle Cloud)**:
   - Implemented in `internal/adapter/repository/oracle/objectstorage`.
   - Uploads files to OCI Object Storage bucket.
   - Public asset endpoint `GET /api/v1/assets/*` dynamically generates an OCI Pre-Authenticated Request (PAR) with 1-hour expiration and returns a `302 Found` redirect.
3. **Asset Constraints**:
   - Only image MIME types (`image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/avif`, `image/svg+xml`) are accepted.
   - Maximum upload file size: 10 MB.

---

## 7. Testing Strategy & Mocking Guidelines

A full test suite is present in the repository. AI agents must add unit and integration tests for every new feature or modification.

### Running Tests

```bash
# Run all tests
task test

# Run unit tests only
task test:unit

# Run HTTP integration tests
task test:integration

# Run tests with race detection & coverage profile
task test:coverage
```

### Test Organization

- **Service Layer Tests** (`internal/service/service_test.go`):
  - Uses mock ports defined in `internal/service/mocks_test.go`.
  - Tests validation rules, business logic errors, and repository interactions.
- **In-Memory Adapter Tests** (`internal/adapter/repository/memory/memory_test.go`):
  - Validates in-memory store functionality used for mock and DAST environments.
- **HTTP Handler Tests** (`internal/adapter/handler/handler_test.go`):
  - Validates request binding, status codes, and DTO transformations.
- **Router Integration Tests** (`internal/router/integration_test.go`):
  - Uses `app.Test(req)` with the in-memory mock store to assert complete HTTP request/response lifecycles, authentication flows, and headers.

---

## 8. Database Migrations Workflow

Migrations are managed with `golang-migrate` across two isolated sequences:

- **ATP (OLTP)**: `migrations/atp/NNNN_name.up.sql` and `NNNN_name.down.sql`
- **ADW (OLAP)**: `migrations/adw/NNNN_name.up.sql` and `NNNN_name.down.sql`

### Migration Rules

1. **NEVER edit existing applied migration files**. Always create a new numbered pair (e.g. `0003_add_feature.up.sql` and `0003_add_feature.down.sql`).
2. Schema types in Oracle DDL:
   - Primary Keys: `VARCHAR2(36 CHAR)` for UUIDs.
   - Enums: `VARCHAR2(20)` with `CHECK (status IN ('draft', 'published', 'archived'))`.
   - Booleans: `NUMBER(1)` with `CHECK (is_featured IN (0, 1))`.
   - JSON Data: `VARCHAR2(4000)` with `CHECK (tech_stack IS JSON)`.
   - Long Text: `CLOB`.

---

## 9. DevSecOps & Security Tools Reference

All security tools are wired into `Taskfile.yml` and GitHub Actions:

```bash
task scan:secrets     # Gitleaks: Scans for committed secrets & API keys
task scan:sast        # Gosec: Static analysis for security vulnerabilities in Go
task scan:sca         # Govulncheck: Audits dependencies for known vulnerabilities
task scan:zizmor      # Zizmor: Audits GitHub Actions workflows for security flaws
task scan:docker      # Trivy: Container vulnerability scanning
task scan:dast        # OWASP ZAP: Dynamic OpenAPI security fuzzing
task scan:all         # Runs all local static scans
```

---

## 10. Agent Quality & Verification Checklist

Before reporting completion on any task, you MUST perform the following checks:

1. **Formatting & Static Analysis**:
   ```bash
   task fmt
   task vet
   task lint
   ```
2. **Test Suite Verification**:
   ```bash
   task test
   ```
3. **Type Port Assertions**:
   - Ensure all new adapters implement their respective interfaces with `var _ port.X = (*Y)(nil)` in `assertions.go`.
4. **Clean Git Status**:
   - Do not leave untracked debug binaries, temporary test files, or hardcoded secrets in the workspace.
