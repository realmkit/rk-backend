# AGENTS.md

This file describes how coding agents must work in the RealmKit Go backend repository.

## Project Overview

RealmKit is a configurable game community forum and platform. It supports user systems, forums, moderation sanctions, appeals, staff pages, friends, messages, statistics, and game-specific integrations such as Minecraft minigames and inventories or SAMP money, inventories, and factions.

This repository contains only the Go backend:

- `pkg/` contains reusable project-level infrastructure.
- `module/` contains business and application-specific modules.
- `plan/` is reserved for local planning notes and must not be committed.

The Next.js frontend lives in the separate `realmkit-frontend` repository.

## Global Rules

- Use Go style, Unix style, and idiomatic formatting.
- Keep code easy to read, professional, reusable, and focused on single responsibility.
- Avoid registry-style types or global service containers unless they are clearly necessary.
- Do not create `catalog.go`, `registry.go`, or similar central list files that collect unrelated module definitions, events, jobs, permissions, routes, handlers, or schemas. Keep definitions owned by the feature/package that produces them, and compose them only at the application composition root when wiring is required.
- Prefer small, explicit interfaces owned by the consumer package.
- Keep interfaces, services, and files focused on one concern. Do not create god interfaces or god services that mix unrelated workflows such as structure admin, content, interactions, operations, and configuration.
- Keep production Go files at or below 250 lines whenever practical. If a package needs more than 6 production files or cannot stay readable within that limit, split it into smaller concern-owned packages.
- Keep struct literals, function calls, and return values readable. Use multi-line composite literals with one field per line when a literal has several fields or would create an excessively long line.
- Keep line length humane. Prefer extracting small helpers or using multi-line formatting over horizontal scrolling.
- Keep behavior idempotent wherever a command, job, migration, or integration can be retried.
- Build for fault tolerance, resilience, and traceability from the start.
- Do not add implementation code until the relevant package/module boundaries are clear.

## Backend Rules

- The backend is a modular, decoupled monolith written in Go.
- Use domain-driven design and hexagonal architecture for every business module.
- Place reusable project-level infrastructure in `pkg`.
- Place business and application-specific modules in `module`.
- Keep application entrypoints in `pkg/cmd`.
- Keep OpenAPI and Swagger specifications in `pkg/api/openapi`.
- Keep database migrations in `pkg/postgres/migrations`.
- Database migrations must be managed as one global PostgreSQL timeline; modules must not run their own independent migration streams.
- Do not rely on GORM `AutoMigrate` for production schema management.
- Each Go package may contain at most 6 production files and their corresponding test files. If a package needs more production files, split it into smaller packages.
- All packages, types, interfaces, functions, methods, constants, and variables must have Go doc comments, including private functions and methods.
- Do not place explanatory comments inside function bodies. Prefer clear names and small functions.
- Internal comments are allowed only for generated code markers or unavoidable tool directives.
- Use Viper for environment-backed configuration.
- Every package that owns configurable behavior must define its own configuration struct in that package.
- The root runtime configuration must compose package-owned configuration structs instead of redefining their fields.
- Config fields must be grouped into purpose-specific structs, such as runtime, PostgreSQL, Redis, and telemetry structs.
- Config fields must use a `default` struct tag when optional. Fields without a `default` tag are mandatory and loading configuration must fail when they are absent.
- Only essential runtime configuration should be environment-driven, such as database credentials, Redis credentials, server bind address, and secrets.
- Do not expose low-level tuning knobs such as circuit breaker delays, pool sizes, retry intervals, or cache internals as required environment variables unless there is a strong operational reason.
- Use PostgreSQL as the primary durable database.
- Use GORM for database access.
- Use Redis for caching, rate limiting, and distributed coordination when appropriate.
- Use Zap for structured JSON logging, with the log level controlled by `REALMKIT_LOG_LEVEL`.
- Use Fiber for HTTP serving and FiberZap for structured request logging.
- Startup logging must use Zap and must emit the RealmKit startup message in every environment.
- Fiber's startup/welcome message must be enabled only in the `development` environment.
- Entrypoints must centralize final error handling through one deferred finalizer that logs with Zap, syncs the logger, and exits nonzero when an error is present.
- Entrypoints must return errors explicitly from setup steps instead of repeating fatal logging at each dependency initialization site.
- Define HTTP contracts through OpenAPI before or alongside handlers.
- Use OpenAPI 3.1 for HTTP contracts.
- RealmKit service routes must not include public version prefixes such as `/api/v1`; public API versioning belongs at the API gateway, which rewrites versioned external URLs to unversioned service routes.
- Every Fiber route must have a corresponding OpenAPI operation before it is considered complete.
- Every OpenAPI operation must document request headers, path/query parameters, request bodies, response headers, success responses, and all expected error responses.
- Every OpenAPI operation must document authentication, authorization, idempotency behavior, rate limit behavior, pagination behavior, and concurrency headers when applicable.
- Every error response must use the shared problem response format and must map to explicit HTTP status codes.
- Every mutating route that can be retried must define whether `Idempotency-Key` is required, optional, or unsupported.
- Rate-limited routes must document `429` plus `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset`, and `Retry-After` when applicable.
- Swagger UI must serve the same OpenAPI contract used by tests and must be enabled by default only in development until production access control exists.
- Fiber HTTP adapters must be thin transport layers that map DTOs to application services and must not contain business rules.
- Design cache behavior deliberately with clear ownership, invalidation, TTL strategy, and failure behavior.
- Include circuit breakers, rate limiters, timeouts, retries, and health checks where integration risk justifies them.
- Make telemetry a first-class requirement: logs, traces, metrics, correlation IDs, and error context must be part of the design.

## Package Design Guidance

- `pkg` is for reusable infrastructure and project-level dependencies such as application entrypoints, API contracts, configuration, migrations, PostgreSQL, Redis, telemetry, HTTP middleware, OpenAPI helpers, and resilience utilities.
- `module` is for bounded contexts such as forum, posts, likes, sanctions, appeals, staff, friends, messages, users, and game integrations.
- Modules should keep domain models independent from transport and persistence concerns.
- Persistence adapters should not leak GORM models into domain APIs.
- Transport adapters should not contain business rules.
- Application services should coordinate use cases without becoming god objects.

## Testing Guidance

- Add focused tests with new implementation work.
- Keep test coverage as high as practical for the changed package, aiming for the maximum meaningful coverage without brittle tests.
- Unit test functions must have Go doc comments just like exported production signatures.
- Keep tests near the package they verify.
- Prefer deterministic unit tests for domain behavior.
- Use integration tests for database, Redis, HTTP contract, and adapter behavior when the boundary matters.
- Tests must be readable and must not depend on hidden global state.

## Git Hygiene

- Do not commit `plan/`.
- Do not commit generated build outputs, local environment files, dependency caches, or editor metadata.
- Do not overwrite user changes. If existing changes conflict with the requested work, inspect them and adapt.
