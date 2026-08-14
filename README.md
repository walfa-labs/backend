# Walfa Labs - Backend API

[![Go Version](https://img.shields.io/badge/Go-1.26.6-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Fiber Framework](https://img.shields.io/badge/Fiber-v3.4.0-00ACD7?style=flat&logo=go&logoColor=white)](https://gofiber.io)
[![Database](https://img.shields.io/badge/Oracle-ATP%20%2B%20ADW%20%2B%2023ai-F80000?style=flat&logo=oracle&logoColor=white)](https://www.oracle.com/cloud/database/)
[![Security](https://img.shields.io/badge/DevSecOps-SAST%20%7C%20DAST%20%7C%20SCA-brightgreen?style=flat&logo=securityscorecard&logoColor=white)](https://github.com/walfa-labs/backend/actions)
[![Docker](https://img.shields.io/badge/Docker-Production_Ready-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![License: WTFPL](https://img.shields.io/badge/License-WTFPL-brightgreen.svg?style=flat)](LICENSE)

A high-performance, enterprise-grade REST API backend for developer portfolios and dynamic content management. Built in Go with the **Fiber v3** framework (fasthttp core), **Sonic JSON** SIMD/JIT parsing, and a clean **Hexagonal Architecture** (Ports & Adapters).

Persistence is designed for **polyglot Oracle Cloud** environments (OCI Always Free tier) while providing seamless offline local development with **Oracle Free 23ai in Docker** and **local disk storage**.

Connected to the [Walfa Labs Nuxt 4 Frontend](../frontend/README.md).

---

## Table of Contents

- [Architecture & Design](#architecture--design)
- [Key Features](#key-features)
- [Tech Stack](#tech-stack)
- [Project Directory Structure](#project-directory-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Option A: Local Development with Docker (Recommended)](#option-a-local-development-with-docker-recommended)
  - [Option B: Oracle Cloud Infrastructure (OCI Production)](#option-b-oracle-cloud-infrastructure-oci-production)
  - [Option C: Zero-Infra Mock / DAST Mode](#option-c-zero-infra-mock--dast-mode)
- [API Reference & Endpoints](#api-reference--endpoints)
  - [Public Endpoints](#public-endpoints-read-only-cached)
  - [Authentication Endpoints](#authentication-endpoints)
  - [Admin Management Endpoints](#admin-management-endpoints-jwt-required)
  - [Analytics & Stats Endpoints](#analytics--stats-endpoints)
  - [System & Documentation](#system--documentation)
- [Security & DevSecOps](#security--devsecops)
- [Configuration Reference](#configuration-reference)
- [Automation & Taskfile](#automation--taskfile)
- [Testing & Quality Assurance](#testing--quality-assurance)
- [Production Deployment](#production-deployment)
- [License](#license)

---

## Architecture & Design

The application follows the **Hexagonal / Clean Architecture** (Ports & Adapters) paradigm. Dependency flow points strictly inward: `adapter` &rarr; `service` &rarr; `port` &rarr; `domain`. The core `domain` package maintains zero internal or external framework dependencies.

```
                               ┌───────────────────────────────────────────────────────┐
                               │                     HTTP CLIENTS                      │
                               └──────────────────────────┬────────────────────────────┘
                                                          │
                                                          ▼
                               ┌───────────────────────────────────────────────────────┐
                               │              Driving Adapter (Fiber v3)               │
                               │   • Router, Middleware (Auth, CORS, Security, Log)    │
                               │   • Request DTOs, Input Validation, Envelopes         │
                               └──────────────────────────┬────────────────────────────┘
                                                          │
                                                          ▼
                               ┌───────────────────────────────────────────────────────┐
                               │                 Application Services                  │
                               │   • Experience, Project, Post, Auth, Asset, Profile   │
                               │   • Business rules, State transitions, Dual-write     │
                               └──────────────────────────┬────────────────────────────┘
                                                          │
                                                          ▼
                               ┌───────────────────────────────────────────────────────┐
                               │                      Port Layer                       │
                               │   • Interfaces: *Repo, AnalyticsStore, AssetStore     │
                               └──────────────────────────┬────────────────────────────┘
                                                          │
                                     ┌────────────────────┴────────────────────┐
                                     │                                         │
                                     ▼                                         ▼
                 ┌───────────────────────────────────────┐ ┌───────────────────────────────────────┐
                 │    Driven Adapters (Oracle / Local)   │ │             Domain Layer              │
                 │ • ATP Repo (godror SQL - OLTP)        │ │ • Core Entities                       │
                 │ • ADW Analytics Store (godror - OLAP) │ │ • Sentinel Errors                     │
                 │ • OCI Object Storage / Local Storage  │ │ • Domain Enums                        │
                 │ • In-Memory Mock Store (Tests & DAST) │ │ (Zero external dependencies)          │
                 └───────────────────────────────────────┘ └───────────────────────────────────────┘
```

---

## Key Features

- ⚡ **Blazing Fast Throughput**: Built on top of Fiber v3 (fasthttp) and ByteDance Sonic JIT/SIMD JSON encoder/decoder.
- 🏛️ **Hexagonal Architecture**: Clear domain boundaries, dependency inversion, and fully swappable persistence adapters.
- 💾 **Polyglot Oracle Persistence**:
  - **OLTP**: Autonomous Transaction Processing (ATP) or Oracle Database 23ai Free for core operational domain models (experiences, projects, blog posts, tags, assets, singleton profile, admin auth).
  - **OLAP / Analytics**: Autonomous Data Warehouse (ADW) or Oracle Database 23ai Free for star-schema analytics (`dim_posts`, `fact_post_views`), powering real-time time-series views and engagement ranking.
  - **Dual Asset Storage**: Cloud-native OCI Object Storage with Pre-Authenticated Request (PAR) generation, or local filesystem storage with static HTTP routing.
- 🔄 **Dual-Write View Tracking**: Non-blocking asynchronous dual-write pattern on public post reads (ATP view counter + ADW analytics fact records).
- 🛡️ **Hardened DevSecOps**:
  - JWT authentication (HS256) with short-lived access tokens (15m) and sliding refresh tokens (7d) via `httpOnly`, `Secure`, `SameSite=Strict` cookies.
  - Rate-limited auth endpoints (5 requests/minute per client IP).
  - Centralized security headers middleware (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy).
  - Full automated security suite: Gitleaks, TruffleHog, Gosec SAST, Semgrep, Govulncheck SCA, Trivy (FS + Container), Zizmor workflow auditing, CodeQL, and OWASP ZAP DAST.
- 🛠️ **Developer Experience & Tooling**:
  - Live OpenAPI 3 specification and interactive Swagger UI at `/swagger`.
  - Structured logging with `gookit/slog` and automatic log file rotation.
  - Standardized JSON success envelopes and RFC-style error envelopes with request correlation ID tracking (`X-Request-ID`).
  - In-Memory Mock repository implementation for instant, zero-infra local development and testing.

---

## Tech Stack

| Component | Technology | Description |
|---|---|---|
| **Language** | [Go 1.26.6](https://go.dev) | High-concurrency compiled backend language |
| **HTTP Engine** | [Fiber v3](https://github.com/gofiber/fiber/v3) (`v3.4.0`) | Express-inspired fasthttp web framework |
| **JSON Codec** | [Sonic](https://github.com/bytedance/sonic) | SIMD/JIT accelerated JSON serializer/deserializer |
| **Database Driver** | [godror](https://github.com/godror/godror) | Native Oracle Database driver via ODPI-C |
| **Persistence (OLTP)** | Oracle ATP / Oracle 23ai Free | Operational transactional data store |
| **Persistence (OLAP)** | Oracle ADW / Oracle 23ai Free | Analytics and star-schema time-series warehouse |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Database schema migrations with Oracle driver |
| **Authentication** | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) + [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | JWT token signing, verification & password hashing |
| **Asset Storage** | [OCI Go SDK](https://github.com/oracle/oci-go-sdk) / Local Disk | Oracle Cloud Object Storage (PAR) & Local filesystem |
| **Validation** | [go-playground/validator/v10](https://github.com/go-playground/validator/v10) | Struct validation mapped via Fiber validator |
| **Logging** | [gookit/slog](https://github.com/gookit/slog) + [rotatefile](https://github.com/gookit/rotatefile) | Structured console and rotating file logger |
| **API Documentation** | [Swagger UI](https://github.com/gofiber/contrib/v3/swaggerui) | Interactive OpenAPI 3 explorer at `/swagger` |
| **Task Automation** | [Task](https://taskfile.dev) | Cross-platform task execution engine (`Taskfile.yml`) |

---

## Project Directory Structure

```
├── .github/
│   └── workflows/              # CI/CD, DevSecOps, DAST, and Release pipelines
│       ├── ci.yml              # Linting, unit/integration testing, build & Trivy scan
│       ├── dast.yml            # OWASP ZAP dynamic security scan against live API
│       ├── release.yml         # Multi-arch container build, SBOM & GitHub Release
│       └── security.yml       # Gitleaks, TruffleHog, Gosec, Semgrep, Govulncheck, Zizmor, CodeQL
├── cmd/
│   └── api/
│       └── main.go             # Application entrypoint: DI wiring, shutdown lifecycle
├── docker/
│   └── init-oracle.sql         # Local Oracle 23ai database initialization script
├── docs/
│   └── openapi.yaml            # OpenAPI 3.0 API specification
├── internal/
│   ├── adapter/
│   │   ├── handler/            # HTTP driving adapters (Fiber handlers, DTO mapping)
│   │   ├── middleware/         # Auth, CORS, Logger, Recover, RequestID, SecurityHeaders
│   │   └── repository/         # Driven persistence adapters
│   │       ├── localstorage/   # Local disk asset storage implementation
│   │       ├── memory/         # In-memory repository & mock store (tests & DAST)
│   │       └── oracle/
│   │           ├── adw/        # Oracle ADW analytics implementation (star schema)
│   │           ├── atp/        # Oracle ATP OLTP repository implementation
│   │           └── objectstorage/# OCI Object Storage adapter (PAR generation)
│   ├── config/                 # Environment configuration loader (caarlos0/env)
│   ├── domain/                 # Core entities, enums, and sentinel errors
│   ├── platform/               # Infrastructure factories (Fiber server, Oracle DB pool, logger)
│   ├── port/                   # Core repository and service interface contracts
│   ├── router/                 # Route registration and endpoint group definitions
│   └── service/                # Business logic and use-case implementation
├── migrations/
│   ├── adw/                    # golang-migrate SQL files for analytics database
│   ├── atp/                    # golang-migrate SQL files for operational database
│   ├── seed.sql                # ATP seed dataset for demo/development
│   └── seed_ora.sql            # Oracle SQLcl-compatible seed script
├── .air.toml                   # Hot reload configuration for Air
├── .gitleaks.toml              # Gitleaks secret scanning rules and allowlists
├── .golangci.yml               # Linter configuration for golangci-lint
├── docker-compose.yml          # Local container stack (Oracle Free 23ai, Backend, ZAP DAST)
├── Dockerfile                  # Multi-stage container build (Oracle Linux 9 + Instant Client)
├── go.mod                      # Go module definition (Go 1.26.6)
└── Taskfile.yml                # Taskfile automation runner
```

---

## Getting Started

### Prerequisites

- **Go**: `1.26.6` or later installed
- **CGO Toolchain**: C compiler required for godror / ODPI-C compilation (`gcc`, `clang`, or `zig cc`)
- **Task Runner**: [Taskfile](https://taskfile.dev/installation/) (`task`)
- **Docker & Docker Compose**: (Required for local Oracle container development)
- **Oracle Instant Client**: (Runtime requirement for local bare-metal execution; built-in for Docker)

---

### Option A: Local Development with Docker (Recommended)

This mode runs **Oracle Database 23ai Free** in Docker and stores asset uploads on your local disk.

#### 1. Setup Environment Configuration
```bash
cp .env.example .env
```

Ensure your `.env` contains the default local settings:
```env
APP_ENV=development
APP_PORT=:8080
STORAGE_DRIVER=local
STORAGE_LOCAL_DIR=./uploads
STORAGE_BASE_URL=http://localhost:8080/uploads
ATP_DSN=portfolio_atp/devpassword@localhost:1521/FREEPDB1
ADW_DSN=portfolio_adw/devpassword@localhost:1521/FREEPDB1
MIGRATE_ATP_URL=oracle://portfolio_atp:devpassword@localhost:1521/FREEPDB1
MIGRATE_ADW_URL=oracle://portfolio_adw:devpassword@localhost:1521/FREEPDB1
JWT_SECRET=dev-secret-change-me
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH='$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
```

#### 2. Start the Oracle Database Container
```bash
# Start Oracle 23ai Free container
task docker:compose-up
```
*Note: Wait ~20–30 seconds on initial startup until the container healthcheck marks healthy.*

#### 3. Install CLI Tooling & Apply Migrations
```bash
# Install development and migration tools (Air, golang-migrate with Oracle support)
task tools

# Apply migrations to ATP and ADW schemas
task migrate-up

# (Optional) Seed initial demo data
task seed
```

#### 4. Run the API Server
```bash
# Option 1: Run with hot-reload via Air
task dev

# Option 2: Standard Go run
task run
```

Access Swagger UI documentation at: **[http://localhost:8080/swagger](http://localhost:8080/swagger)**

---

### Option B: Oracle Cloud Infrastructure (OCI Production)

For deployment against live Oracle Cloud Autonomous Databases and OCI Object Storage:

1. **Obtain OCI Credentials & Wallet**:
   - Download ATP & ADW client credentials wallet `.zip` files and extract them to `./wallet`.
   - Create an OCI Object Storage bucket (e.g., `portfolio-assets`).
   - Generate an OCI API Signing Key pair (`oci_api_key.pem`).
2. **Configure `.env`**:
   ```env
   APP_ENV=production
   APP_PORT=:8080
   TNS_ADMIN=./wallet
   ATP_DSN=portfolio/your_password@portfolio_atp_high
   ADW_DSN=portfolio/your_password@portfolio_adw_high
   STORAGE_DRIVER=oci
   OCI_TENANCY_OCID=ocid1.tenancy.oc1..your_tenancy
   OCI_USER_OCID=ocid1.user.oc1..your_user
   OCI_FINGERPRINT=aa:bb:cc:...
   OCI_REGION=ap-singapore-1
   OCI_PRIVATE_KEY_PATH=./oci_api_key.pem
   OCI_BUCKET=portfolio-assets
   ```
3. **Migrate & Start**:
   ```bash
   task migrate-up
   task run
   ```

---

### Option C: Zero-Infra Mock / DAST Mode

Run the entire application in-memory with zero external database or storage dependencies:

```bash
APP_ENV=mock ATP_DSN=mock ADW_DSN=mock JWT_SECRET=mock-secret ADMIN_PASSWORD_HASH='$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' go run ./cmd/api
```

---

## API Reference & Endpoints

Interactive documentation with request/response schemas is available via Swagger UI at `/swagger`.

### Public Endpoints (Read-Only, Cached)

| Method | Endpoint | Description | Cache / Behavior |
|---|---|---|---|
| `GET` | `/api/v1/health` | Service healthcheck (pings ATP & ADW) | `no-store` |
| `GET` | `/api/v1/experiences` | List professional work experiences | `public, max-age=300` |
| `GET` | `/api/v1/experiences/:id` | Get specific experience details | `public, max-age=300` |
| `GET` | `/api/v1/projects` | List published projects (supports `?featured=true`) | `public, max-age=300` |
| `GET` | `/api/v1/projects/:slug` | Get project by URL slug | `public, max-age=300` |
| `GET` | `/api/v1/blog/posts` | List published blog posts (paginated, `?tag=`) | `public, max-age=180` |
| `GET` | `/api/v1/blog/posts/:slug` | Get post by slug (increments views in ATP & ADW) | `public, max-age=180` |
| `GET` | `/api/v1/tags` | List all unique blog post tags with post counts | `public, max-age=600` |
| `GET` | `/api/v1/stats/summary` | Portfolio summary statistics | `public, max-age=300` |
| `GET` | `/api/v1/profile` | Public singleton portfolio profile | `public, max-age=600` |
| `GET` | `/api/v1/assets/*` | Resolve asset key & redirect (302) to signed PAR URL | `302 Found` |
| `GET` | `/uploads/*` | Static file serving (only active when `STORAGE_DRIVER=local`) | Direct File |

### Authentication Endpoints

| Method | Endpoint | Description | Auth / Rate Limit |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | Authenticate admin; returns JWT & sets refresh cookie | 5 req/min per IP |
| `POST` | `/api/v1/auth/refresh` | Refresh expired access token using refresh token | Token required |

### Admin Management Endpoints (JWT Required)

All admin routes require a valid Bearer token (`Authorization: Bearer <token>`).

| Resource | Method | Endpoint | Description |
|---|---|---|---|
| **Experiences** | `GET` | `/api/v1/admin/experiences` | List all experiences |
| | `GET` | `/api/v1/admin/experiences/:id` | Get experience by ID |
| | `POST` | `/api/v1/admin/experiences` | Create new experience |
| | `PUT/PATCH` | `/api/v1/admin/experiences/:id` | Update experience |
| | `DELETE` | `/api/v1/admin/experiences/:id` | Delete experience |
| **Projects** | `GET` | `/api/v1/admin/projects` | List all projects (including drafts) |
| | `GET` | `/api/v1/admin/projects/:id` | Get project by ID |
| | `POST` | `/api/v1/admin/projects` | Create new project |
| | `PUT/PATCH` | `/api/v1/admin/projects/:id` | Update project |
| | `DELETE` | `/api/v1/admin/projects/:id` | Delete project |
| **Blog Posts** | `GET` | `/api/v1/admin/blog/posts` | List all posts (drafts, archived, published) |
| | `GET` | `/api/v1/admin/blog/posts/:id` | Get blog post by ID |
| | `POST` | `/api/v1/admin/blog/posts` | Create new blog post |
| | `PUT/PATCH` | `/api/v1/admin/blog/posts/:id` | Update blog post |
| | `DELETE` | `/api/v1/admin/blog/posts/:id` | Delete blog post |
| | `PATCH` | `/api/v1/admin/blog/posts/:id/status` | Transition post status (`draft`, `published`, `archived`) |
| **Assets** | `POST` | `/api/v1/admin/assets` | Upload image asset (multipart, max 10MB) |
| | `DELETE` | `/api/v1/admin/assets/*` | Delete asset by key |
| **Profile** | `GET` | `/api/v1/admin/profile` | Get admin singleton profile details |
| | `PUT` | `/api/v1/admin/profile` | Upsert singleton profile details |

### Analytics & Stats Endpoints

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api/v1/admin/stats/views` | Time-series view analytics (`?from=&to=&bucket=day\|week\|month`) | Admin JWT |
| `GET` | `/api/v1/admin/stats/top-posts` | Top performing blog posts ranking (`?limit=10`) | Admin JWT |

### System & Documentation

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/swagger` | Interactive Swagger UI API Explorer |
| `GET` | `/docs/openapi.yaml` | Raw OpenAPI 3.0 YAML specification |

---

## Security & DevSecOps

This repository integrates enterprise-grade DevSecOps security controls natively in CI/CD pipelines and local task runners:

| Security Domain | Tooling | Execution Phase | Description |
|---|---|---|---|
| **Secret Detection** | Gitleaks, TruffleHog | Pre-commit & CI Pipeline | Scans commits and history for leaked secrets & credentials |
| **SAST (Go Code)** | Gosec, Semgrep | CI Pipeline & Local Task | Static AST analysis for Go security vulnerabilities & OWASP rules |
| **SAST (Deep)** | GitHub CodeQL | Scheduled & CI Pipeline | Deep semantic query scanning for tainted inputs & injection flaws |
| **SCA (Dependencies)** | Govulncheck, Trivy | CI Pipeline & Local Task | Audits Go modules & dependencies against known CVE databases |
| **Workflow Audit** | Zizmor | CI Security Scan | Static analysis of GitHub Actions workflows for security misconfigurations |
| **Container Scan** | Trivy Container | Docker Build Step | Scans base OS images and container packages for CVEs |
| **DAST (Dynamic)** | OWASP ZAP (OpenAPI) | Pull Requests & Weekly Schedule | Dynamic security testing against live API using OpenAPI definitions |
| **Supply Chain** | CycloneDX SBOM | Tagged Releases to GHCR | Generates Software Bill of Materials (SBOM) on release |

Run all static security scans locally with:
```bash
task scan:all
```

---

## Configuration Reference

All settings are read from environment variables via `internal/config`:

| Variable | Type | Default | Description |
|---|---|---|---|
| `APP_ENV` | `string` | `development` | Runtime environment (`development`, `production`, `mock`, `dast`) |
| `APP_PORT` | `string` | `:8080` | TCP port for HTTP server binding |
| `ATP_DSN` | `string` | *(Required)* | godror connect string for OLTP ATP database (`user/pass@host:port/service`) |
| `ADW_DSN` | `string` | *(Required)* | godror connect string for OLAP ADW database (`user/pass@host:port/service`) |
| `JWT_SECRET` | `string` | *(Required)* | HMAC-SHA256 secret key for signing JWT tokens |
| `JWT_ACCESS_TTL` | `duration` | `15m` | Lifetime duration for access tokens |
| `JWT_REFRESH_TTL` | `duration` | `168h` | Lifetime duration for refresh tokens (7 days) |
| `ADMIN_USERNAME` | `string` | `admin` | Default admin username |
| `ADMIN_PASSWORD_HASH` | `string` | *(Required)* | Bcrypt password hash for admin login |
| `STORAGE_DRIVER` | `string` | `local` | Asset backend: `local` (disk) or `oci` (Oracle Object Storage) |
| `STORAGE_LOCAL_DIR` | `string` | `./uploads` | Local directory for storing asset uploads |
| `STORAGE_BASE_URL` | `string` | `http://localhost:8080/uploads` | Base URL used for resolving local asset URLs |
| `OCI_TENANCY_OCID` | `string` | `""` | OCI Tenancy OCID (required if `STORAGE_DRIVER=oci`) |
| `OCI_USER_OCID` | `string` | `""` | OCI User OCID (required if `STORAGE_DRIVER=oci`) |
| `OCI_FINGERPRINT` | `string` | `""` | OCI Public key fingerprint |
| `OCI_REGION` | `string` | `""` | OCI Region identifier (e.g., `ap-singapore-1`) |
| `OCI_PRIVATE_KEY_PATH` | `string` | `""` | Absolute or relative path to OCI private key PEM file |
| `OCI_NAMESPACE` | `string` | `""` | OCI Object Storage namespace (auto-resolved if omitted) |
| `OCI_BUCKET` | `string` | `""` | Target OCI Object Storage bucket name |
| `CORS_ALLOWED_ORIGINS` | `[]string` | `http://localhost:3000` | Comma-separated allowlist of CORS origins |
| `MIGRATE_ATP_URL` | `string` | - | Used by `task migrate-up` for ATP migrations |
| `MIGRATE_ADW_URL` | `string` | - | Used by `task migrate-up` for ADW migrations |

---

## Automation & Taskfile

This repository includes a unified [`Taskfile.yml`](Taskfile.yml) for developer operations. Run `task --list` to view all commands:

### Development & Build
```bash
task run              # Start API server locally (loads .env)
task dev              # Run server with hot reload via Air
task build            # Compile production binary into bin/
task clean            # Remove build artifacts, coverage files, and temporary directories
task tidy             # Sync and tidy go.mod and go.sum dependencies
task fmt              # Format all Go source files
task vet              # Run Go static analysis (go vet)
task tools            # Install development and security tools
```

### Testing & Code Quality
```bash
task test             # Run all unit and integration tests
task test:unit        # Run unit tests only
task test:integration # Run HTTP route integration tests
task test:coverage    # Run test suite with race detector and coverage profiling
task test:html        # Generate an HTML code coverage report
task lint             # Run golangci-lint
task lint:fix         # Run golangci-lint and auto-fix supported issues
```

### Security Scans (DevSecOps)
```bash
task scan:all         # Run all local static security scans
task scan:secrets     # Run Gitleaks secret detection
task scan:sast        # Run Gosec SAST analyzer
task scan:sca         # Run Govulncheck dependency vulnerability scanner
task scan:zizmor      # Audit GitHub Actions workflows for vulnerabilities
task scan:docker      # Scan container image with Trivy
task scan:dast        # Run OWASP ZAP DAST container against live API
```

### Database & Docker
```bash
task migrate-up       # Apply all database migrations to ATP and ADW
task migrate-down     # Roll back all database migrations
task seed             # Load demo dataset into ATP via SQLcl
task docker:build     # Build local Docker container image
task docker:run       # Run local Docker container
task docker:stop      # Stop and remove local Docker container
task docker:compose-up   # Start local stack with Oracle 23ai Free
task docker:compose-down # Stop local Docker compose stack
```

---

## Testing & Quality Assurance

The codebase features comprehensive test suites spanning domain, service, repository, middleware, handler, and HTTP integration layers.

To run the entire test suite with race condition detection and coverage reporting:

```bash
task test:coverage
```

### Test Suites
- **Domain Tests** (`internal/domain`): Validates domain validation rules, errors, and entity behaviors.
- **Service Unit Tests** (`internal/service`): Tests business logic, state transitions, and dual-write operations against mock repository ports.
- **Repository Tests** (`internal/adapter/repository/memory`): Verifies CRUD operations and querying in the memory adapter.
- **Handler Tests** (`internal/adapter/handler`): Validates request parsing, JSON deserialization, input validation, and HTTP status codes.
- **Middleware Tests** (`internal/adapter/middleware`): Tests JWT auth verification, error handler envelopes, CORS headers, and security headers.
- **Integration Tests** (`internal/router`): End-to-end HTTP request and response tests using Fiber's `app.Test()` engine.

---

## Production Deployment

### Multi-Stage Container Build

```bash
# Build the production image
docker build -t walfa-labs-backend:latest .

# Run container with environment configuration
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  --name walfa-labs-backend \
  walfa-labs-backend:latest
```

---

## License

Distributed under the WTFPL (Do What The Fuck You Want To Public License). See [LICENSE](LICENSE) for details or visit <http://www.wtfpl.net/>.
