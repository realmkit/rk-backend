# AGENTS.md

This file describes how coding agents must work in the GameHub repository.

## Project Overview

GameHub is a configurable game community forum and platform. It supports user systems, forums, moderation sanctions, appeals, staff pages, friends, messages, statistics, and game-specific integrations such as Minecraft minigames and inventories or SAMP money, inventories, and factions.

The repository is a monorepo with a Go backend and a future Next.js frontend:

- `backend/` contains the Go backend.
- `frontend/` will contain the Next.js frontend.
- `plan/` is reserved for local planning notes and must not be committed.

## Global Rules

- Use Go style, Unix style, and idiomatic formatting.
- Keep code easy to read, professional, reusable, and focused on single responsibility.
- Avoid registry-style types or global service containers unless they are clearly necessary.
- Prefer small, explicit interfaces owned by the consumer package.
- Keep behavior idempotent wherever a command, job, migration, or integration can be retried.
- Build for fault tolerance, resilience, and traceability from the start.
- Do not add implementation code until the relevant package/module boundaries are clear.

## Backend Rules

- The backend is a modular, decoupled monolith written in Go.
- Use domain-driven design and hexagonal architecture for every business module.
- Place reusable project-level infrastructure in `backend/pkg`.
- Place business and application-specific modules in `backend/module`.
- Keep application entrypoints in `backend/pkg/cmd`.
- Keep OpenAPI and Swagger specifications in `backend/pkg/api/openapi`.
- Keep database migrations in `backend/pkg/migrations`.
- Each Go package may contain at most 6 production files and their corresponding test files. If a package needs more production files, split it into smaller packages.
- All exported packages, types, interfaces, functions, methods, and significant constants must have Go doc comments.
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
- Use Zap for structured JSON logging, with the log level controlled by `GAMEHUB_LOG_LEVEL`.
- Define HTTP contracts through OpenAPI before or alongside handlers.
- Design cache behavior deliberately with clear ownership, invalidation, TTL strategy, and failure behavior.
- Include circuit breakers, rate limiters, timeouts, retries, and health checks where integration risk justifies them.
- Make telemetry a first-class requirement: logs, traces, metrics, correlation IDs, and error context must be part of the design.

## Frontend Rules

- The frontend will be a Next.js application under `frontend/`.
- Do not scaffold the Next.js application until requested.
- Keep frontend implementation aligned with the backend contracts and product modules.
- Favor reusable UI primitives and clear domain-facing screens once the frontend is created.

## Package Design Guidance

- `backend/pkg` is for reusable infrastructure and project-level dependencies such as application entrypoints, API contracts, configuration, migrations, PostgreSQL, Redis, telemetry, HTTP middleware, OpenAPI helpers, and resilience utilities.
- `backend/module` is for bounded contexts such as forum, posts, likes, sanctions, appeals, staff, friends, messages, users, and game integrations.
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
