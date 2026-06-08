# GameHub Go

GameHub Go is the backend for GameHub, a configurable game community platform for forums, user accounts, moderation workflows, appeals, staff pages, friends, messages, statistics, and game-specific integrations.

The frontend lives in the separate `gamehub-frontend` repository.

## Backend Direction

The Go backend uses:

- Viper for essential environment-backed configuration.
- PostgreSQL for durable relational storage.
- GORM for database access.
- Redis for caching, rate limiting, and distributed coordination where appropriate.
- Fiber for HTTP serving with FiberZap request logging.
- OpenAPI and Swagger specifications for HTTP contracts.
- Telemetry, tracing, metrics, and structured logs as first-class concerns.

The module path is `github.com/niflaot/gamehub-go`.

Runtime configuration is loaded by `pkg/config` through Viper using `.env` files and `GAMEHUB_` environment variables:

- `GAMEHUB_HOST`, default `0.0.0.0`
- `GAMEHUB_PORT`, default `8080`
- `GAMEHUB_ENVIRONMENT`, default `development`
- `GAMEHUB_LOG_LEVEL`, default `info`
- `GAMEHUB_POSTGRES_HOST`, default `localhost`
- `GAMEHUB_POSTGRES_PORT`, default `5432`
- `GAMEHUB_POSTGRES_DATABASE`, required
- `GAMEHUB_POSTGRES_USERNAME`, required
- `GAMEHUB_POSTGRES_PASSWORD`, required
- `GAMEHUB_POSTGRES_SSL_MODE`, default `disable`

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
