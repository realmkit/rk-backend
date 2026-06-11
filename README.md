# RealmKit Go

RealmKit Go is the backend for RealmKit, a configurable game community platform for forums, user accounts, moderation workflows, appeals, staff pages, friends, messages, statistics, and game-specific integrations.

The frontend lives in the separate `realmkit-frontend` repository.

## Backend Direction

The Go backend uses:

- Viper for essential environment-backed configuration.
- PostgreSQL for durable relational storage.
- GORM for database access.
- Redis for caching, rate limiting, and distributed coordination where appropriate.
- Fiber for HTTP serving with FiberZap request logging.
- OpenAPI and Swagger specifications for HTTP contracts.
- Telemetry, tracing, metrics, and structured logs as first-class concerns.

The module path is `github.com/realmkit/rk-backend`.

Runtime configuration is loaded by `pkg/config` through Viper using `.env` files and `REALMKIT_` environment variables:

- `REALMKIT_HOST`, default `0.0.0.0`
- `REALMKIT_PORT`, default `8080`
- `REALMKIT_ENVIRONMENT`, default `development`
- `REALMKIT_LOG_LEVEL`, default `info`
- `REALMKIT_POSTGRES_HOST`, default `localhost`
- `REALMKIT_POSTGRES_PORT`, default `5432`
- `REALMKIT_POSTGRES_DATABASE`, required
- `REALMKIT_POSTGRES_USERNAME`, required
- `REALMKIT_POSTGRES_PASSWORD`, required
- `REALMKIT_POSTGRES_SSL_MODE`, default `disable`

## Repository Layout

```text
.
├── module/             # Application modules and bounded contexts
├── pkg/                # Reusable project-level infrastructure packages
│   ├── api/openapi/    # OpenAPI and Swagger contract files
│   ├── cmd/            # Application entrypoints
│   ├── config/         # Runtime configuration loader
│   └── migrations/     # Database migrations
├── .env.example
├── AGENTS.md
└── README.md
```

## Development

Run tests:

```bash
go test ./...
```

Run the backend:

```bash
go run ./pkg/cmd
```

## API

RealmKit exposes unversioned service routes such as `/forums/tree`, `/assets`, and `/users/me`. Public API versioning is owned by the API gateway, which should publish versioned external paths and rewrite them to the service routes.

OpenAPI is embedded from `pkg/api/openapi/realmkit.v1.json`. In development, Swagger UI is served at `/docs` and the raw OpenAPI contract is served at `/docs/openapi.json`.
