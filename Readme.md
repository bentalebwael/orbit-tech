# 🚀 Student Management System — Technical Submission

## ⚡ Quick Start — Run locally in 60 seconds

- One-command bring-up:

```bash
make start
```

- Generate a PDF report:

```bash
curl -L "http://localhost:8080/api/v1/students/1/report" -o student_1_report.pdf
```

- View logs and stop:

```bash
make logs
make stop
```

This stack is production-leaning yet developer-friendly—robust building blocks, clean separation of concerns, and the right operational guardrails. 💪

---

This document highlights the design choices, architecture decisions, testing strategy, documentation, caching, build and run automation, containerization, error handling, and other strengths implemented across the monorepo.

- Root services: go-service (Golang microservice for PDF reports), frontend (React+Vite), infrastructure (Docker Compose, Makefiles), database (SQL schema + seeders).
- Evidence in repo: docker-compose.yml, Makefile (root and go-service), go-service/internal/*, go-service/api/openapi.yaml, seed_db/*.

## 🧠 Architecture & Design Principles

- Layered, modular architecture
  - Go-service: clean layering with Handler → Service → Cache/External → PDF generator. Interfaces (`internal/service/interfaces.go`, `internal/external/interfaces.go`) enforce low coupling and high testability.
- Integration boundary (no DB coupling)
  - The Go microservice never touches the database. It consumes an upstream REST API via `internal/external/backend.go`.
- Strict boundary at HTTP
  - End-to-end flow: handler validates → service orchestrates → external client fetches → PDF generator renders → cache stores/serves. Changes in one layer don’t ripple across others.
- Graceful runtime behavior
  - Go service initializes structured logging, dependency wiring, and a web server with graceful shutdown (`cmd/api/main.go`).

## 📜 API Design & Documentation

- OpenAPI 3.0 documentation is provided for the Go microservice at `go-service/api/openapi.yaml`:
  - Endpoints: `GET /health`, `GET /api/v1/students/{studentId}/report`.
  - Detailed response shapes and error codes including 400/404/429/500/503, with headers like `X-Request-ID` and `Content-Disposition`.
  - Input validation for `studentId` via regex (1–20 numeric characters), mirrored in handler validation.

## 🧰 Caching Strategy (Performance & Cost)

- File-backed, content-addressed cache in the Go microservice (`internal/cache/file_cache.go`):
  - Keying: SHA-256 hash of student content (e.g., name, class, section, timestamps) created via `GenerateStudentHash` → avoids stale PDFs when the source data changes even if ID remains the same.
  - Storage: PDFs persisted to disk with an in-memory index for fast lookups; old versions for the same student are cleaned on write.
  - TTL-based expiry with a background cleanup worker (runs every minute); expired files removed from disk and index.
  - Optimized happy-path latency: serve-from-cache avoids upstream calls or regenerating PDFs.
  - Failure-tolerant: cache is optional and non-blocking—generation proceeds even if cache read/write fails.

## 🧾 Error Handling & Resilience

- Go microservice uses typed errors (`internal/errors/errors.go`): `NotFoundError`, `ServiceError`, `PDFGenerationError`.
  - Handlers map domain errors to correct HTTP statuses (404, 503, 500) with consistent JSON error envelopes.
  - External client implements retry with exponential backoff (1s, 2s, 4s) for transient upstream failures.

## 🔐 Security Posture

- Service-to-service authentication via `X-API-Key` tied to an environment variable (`INTERNAL_SERVICE_API_KEY`) when communicating with the upstream API.
- Basic hardening middleware in the Go service:
  - Security headers (`X-Content-Type-Options`, `X-Frame-Options`), no-store cache on sensitive routes.
  - Optional per-IP rate limiting (`internal/middleware/rate_limit.go`) with a simple in-memory limiter, configurable via env.

## 📈 Observability & Operability

- Request tracing with `X-Request-ID` propagation in the Go service; every request gets a UUID (`request_id` middleware) and is echoed back in responses.
- Structured, contextual logging using Uber’s Zap across the Go service (request logs, PDF generation, cache hits/misses, retries), ideal for centralized log aggregation.
- Health endpoint:
  - Go service: `/health` probes upstream reachability and exposes an overall status (healthy/degraded).

## ✅ Testing & Coverage

- Go service has comprehensive unit tests across service, cache, and handlers. Coverage artifact `go-service/coverage.out` reports:
  - Total coverage: 93.1% statements (computed via `go tool cover -func coverage.out`).
  - Focused tests on edge cases: caching TTL expiry, concurrent access, upstream failures, invalid IDs, PDF generation errors.
  - Benchmarks included for performance tracking (`make benchmark`).
- Clear testing ergonomics via Make:
  - `make test`, `make test-coverage`, `make coverage-html`, `make test-unit`, plus lint/vet/fmt targets ensure a predictable CI path.

## 🧩 Low Coupling, High Cohesion

- Interface-driven design in the Go service decouples components:
  - `ReportService`, `PDFGenerator`, and `BackendService` interfaces enable mocking and substitution without invasive changes.
  - Handlers don’t know or care if data comes from cache or live upstream—they depend on the service contract.

## 📦 Containerization & Orchestration

- Docker Compose (`docker-compose.yml`) provides a clean multi-container setup for local dev/testing:
  - Networking: services communicate over a private bridge network for internal calls.
  - Health checks: the Go service exposes `/health` used by Docker to monitor liveness.
- Go service Dockerfile uses a multi-stage build for small final images (Alpine base with CA certs) and an internal health check.

## 🛠️ Makefiles that Supercharge DX

- Root `Makefile` provides one-command onboarding:
  - `make start` builds images, starts containers, waits for database readiness, and seeds the database automatically from `seed_db/*.sql`.
  - `make logs`, `make stop`, `make restart`, `make docker-clean`, `make db-seed` streamline day-to-day ops.
  - Dev helpers: `make go-test`, `make go-fmt`.
- Go service `Makefile` is a complete developer toolbox:
  - Build/run (race), deps management, lint/format/vet, coverage, HTML reports, benchmarks, security scans (`gosec`), Swagger generation, hot reload with `air`.
  - CI-friendly target `make ci` for a compact pipeline.

## 🗃️ Database Schema & Seeders

- Schema (`seed_db/tables.sql`) covers roles, access control, students (users + user_profiles), notices, leaves, departments, classes/sections, and more with foreign keys and unique constraints.
- Seeders (`seed_db/seed-db.sql`) bootstrap an admin and a realistic set of students, classes, and sections to enable immediate end-to-end testing.
- Utility SQL functions (e.g., `student_add_update`, `staff_add_update`) encapsulate upsert-like behaviors, centralizing consistency rules.

## 🖨️ PDF Generation Quality

- Professional, readable PDF layout with clear sections: Personal, Academic, Parents, Guardian, Addresses.
- Consistent typography, headers, and a standardized footer with a generated report ID.
- Robust data formatting helpers: safe handling for empty values, ISO date parsing, and boolean rendering (Active/Inactive).

## ⚙️ Configuration Management

- Environment-driven configuration for the Go service is centralized in `internal/config/config.go` with `envconfig` and `.env` support, including toggles for cache and rate limiting.
- Sensible defaults with overrides to support local dev and containerized deployments.

## 🧭 Quality Gates & Runtime Hygiene

- Build and runtime checks:
  - Docker health check for the Go service.
  - Graceful shutdown handling in the Go service.
- Lint/format/vet in Go ensure codebase consistency and correctness.
- Request/response logging with timings promotes swift incident triage.

## 🌟 What Stands Out

- Thoughtful microservice boundary: the Go service consumes a stable REST contract rather than the DB, enabling independent scaling and evolution.
- Resilience-first design: retries, typed errors, graceful degradation when cache or the upstream API is unavailable.
- DevEx excellence: Makefiles + Docker Compose reduce onboarding to a single command while preserving local flexibility.
- Documentation maturity: OpenAPI spec, README(s), and clear code organization make the service approachable.
- Performance-aware: cached PDFs, content hashing, and targeted cleanup minimize compute and I/O overhead.

## 🔭 Suggested Next Steps (Optional)

- Add end-to-end integration tests against the upstream API (or a contract test/mock server) to complement unit tests.
- Consider serving the OpenAPI spec via Swagger UI for quick manual testing.
- Persist cache metadata to survive restarts (current design intentionally wipes the cache directory on boot for consistency).
- Add distributed tracing headers propagation (e.g., W3C Trace Context) for deeper observability across calls.
