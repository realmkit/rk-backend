# GameHub

GameHub is a configurable game community platform for communities that need forums, user accounts, moderation workflows, appeals, staff pages, friends, messages, statistics, and game-specific integrations.

The platform is designed to support different game ecosystems from one reusable foundation. A Minecraft community might expose minigame stats, inventories, ranks, and server data. A SAMP community might expose money, inventory, factions, punishments, and other domain-specific records.

## Architecture

GameHub is planned as a monorepo with two primary applications:

- `backend/`: Go backend implemented as a modular, decoupled monolith.
- `frontend/`: Next.js frontend application, to be created later.

The backend should follow domain-driven design and hexagonal architecture. Business features live in isolated modules, while shared infrastructure is kept separate and reusable.

## Backend Direction

The Go backend will use:

- Viper for essential environment-backed configuration.
- PostgreSQL for durable relational storage.
- GORM for database access.
- Redis for caching, rate limiting, and distributed coordination where appropriate.
- OpenAPI and Swagger specifications for HTTP contracts.
- Telemetry, tracing, metrics, and structured logs as first-class concerns.

The backend is initialized as the Go module `github.com/niflaot/gamehub/backend`.

Initial runtime configuration is loaded by `backend/pkg/config` through Viper using `.env` files and `GAMEHUB_` environment variables:

- `GAMEHUB_HOST`, default `0.0.0.0`
- `GAMEHUB_PORT`, default `8080`
- `GAMEHUB_ENVIRONMENT`, default `development`

The backend must be designed for idempotency, resilience, and fault tolerance. Cross-cutting behavior such as rate limiting, retries, circuit breakers, timeouts, health checks, and cache policy should be implemented in deliberate infrastructure packages rather than scattered through feature code.

## Repository Layout

```text
.
├── backend/
│   ├── module/         # Application modules and bounded contexts
│   └── pkg/            # Reusable project-level infrastructure packages
│       ├── api/openapi/ # OpenAPI and Swagger contract files
│       ├── cmd/         # Application entrypoints
│       ├── config/      # Configuration examples and documentation
│       └── migrations/  # Database migrations
├── frontend/           # Future Next.js frontend application
├── AGENTS.md           # Project instructions for coding agents
└── README.md
```

## Module Layout

Backend modules should be created under `backend/module/<name>` and should keep domain, application, transport, and persistence responsibilities separate. Shared project-level concerns such as API contracts, entrypoints, configuration, migrations, PostgreSQL, Redis, telemetry, and HTTP middleware belong under `backend/pkg`.

## Current Status

This repository currently contains the base structure, project documentation, and ignore rules only. Backend and frontend implementation work has not started yet.
