# E2E Ecosystem

This directory owns GameHub's end-to-end test ecosystem.

The goal is to test real HTTP journeys through the monolith while keeping setup explicit, small, and readable. The first layer is intentionally local and fast: it starts the real Fiber server in process, uses a migrated in-memory database fixture, uses `miniredis` for Redis-dependent middleware, captures Zap JSON logs, and uses an in-memory S3-compatible storage fixture.

## Run

```sh
go test ./e2e/...
```

## Bootstrap

Shared e2e setup lives in `e2e/harness`.

`harness.New(t)` creates:

- `App`: the real Fiber server under test.
- `Database`: a migrated SQLite in-memory database using the global migration runner.
- `Redis`: an isolated `miniredis` server.
- `RedisClient`: a Redis client connected to the fixture.
- `Storage`: an in-memory implementation of `pkg/storage.Store`.
- `StorageBucket`: the bucket name to pass to storage-backed services.
- `Log`: a Zap logger.
- `LogBuffer`: captured JSON log output for assertions.

The local database fixture is not a substitute for PostgreSQL-specific behavior. It is the fast default for cross-module HTTP journeys. When a scenario needs PostgreSQL fidelity, create a separate suite tier that starts real PostgreSQL through Docker/Testcontainers and reuses the same service wiring style.

The storage fixture is intentionally S3-shaped but in memory. It supports `Put`, `Head`, `Delete`, `PresignPut`, `PresignGet`, and direct `Seed` setup. Do not use local filesystem object storage for e2e tests.

## Dependency Wiring

Add module routes by passing server options:

```go
ecosystem := harness.New(
	t,
	harness.WithServerOptions(
		server.WithAssets(assetServices),
	),
)
```

Use the bootstrap fixtures when composing module services:

```go
assets := assetsapplication.NewService(
	assetRepository,
	ecosystem.Storage,
	ecosystem.StorageBucket,
)
```

Keep each scenario responsible for wiring only the modules it uses. Avoid package-wide service registries, global containers, and hidden init state.

## Naming

Use numeric prefixes for suite order and stable reading order. File and directory names use lower snake case because Go files cannot use hyphen separators.

Recommended root files:

- `00_bootstrap_smoke_test.go`: server and harness smoke checks.
- `01_core_config_test.go`: config/runtime behavior.
- `01_core_redis_test.go`: Redis-backed middleware and cache journeys.
- `01_core_logging_test.go`: request IDs, correlation IDs, and structured logs.

Recommended module suite directories:

- `02_metadata/01_definitions_test.go`
- `03_users/01_registration_test.go`
- `04_groups/01_permissions_test.go`
- `05_assets/01_upload_intent_test.go`
- `06_forums/01_public_tree_test.go`
- `07_punishments/01_restrictions_test.go`
- `08_tickets/01_appeal_flow_test.go`

Package names inside numbered directories should be readable and not numeric:

```go
package forums_e2e
```

## Size Rules

The same readability rules apply here:

- Keep each package at six production Go files or fewer.
- Keep Go files under 250 lines whenever practical.
- Split large suites by module directory instead of growing one root package.
- Keep every helper small and named for the dependency or journey it owns.
- Add doc comments to package-level types, functions, constants, variables, and tests.

If a module suite needs many journeys, split by workflow:

```text
06_forums/
  01_public_tree_test.go
  02_thread_lifecycle_test.go
  03_post_interactions_test.go
```

If that directory approaches the package file limit, create a narrower subpackage:

```text
06_forums/moderation/
  01_hide_restore_test.go
```

## Test Style

Prefer full user journeys over low-level assertions:

1. Build only the services needed by the scenario.
2. Start `harness.New`.
3. Send HTTP requests through `ecosystem.Test`.
4. Assert HTTP status, headers, response body, persisted state, emitted events, and logs when relevant.
5. Keep PostgreSQL-specific query behavior in adapter tests or a future PostgreSQL-fidelity e2e tier.
