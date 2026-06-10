CREATE TABLE event_outbox (
    id uuid PRIMARY KEY,
    event_key text NOT NULL,
    schema_version integer NOT NULL,
    producer text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NULL,
    payload_json jsonb NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}',
    actor_user_id uuid NULL,
    request_id text NOT NULL DEFAULT '',
    correlation_id text NOT NULL DEFAULT '',
    idempotency_key text NOT NULL DEFAULT '',
    dedupe_key text NULL,
    occurred_at timestamptz NOT NULL,
    available_at timestamptz NOT NULL,
    status text NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0,
    locked_by text NOT NULL DEFAULT '',
    locked_until timestamptz NULL,
    processed_at timestamptz NULL,
    dead_at timestamptz NULL,
    last_error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX event_outbox_status_available_idx ON event_outbox (status, available_at, id);
CREATE INDEX event_outbox_producer_key_occurred_idx ON event_outbox (producer, event_key, occurred_at DESC);
CREATE INDEX event_outbox_aggregate_idx ON event_outbox (aggregate_type, aggregate_id, occurred_at DESC);
CREATE INDEX event_outbox_correlation_idx ON event_outbox (correlation_id);
CREATE UNIQUE INDEX event_outbox_dedupe_key_idx ON event_outbox (dedupe_key) WHERE dedupe_key IS NOT NULL;

CREATE TABLE event_scopes (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    scope_type text NOT NULL,
    scope_id text NOT NULL DEFAULT '',
    permission text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX event_scopes_event_idx ON event_scopes (event_id);
CREATE INDEX event_scopes_scope_idx ON event_scopes (scope_type, scope_id, event_id);

CREATE TABLE event_delivery_attempts (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    consumer_key text NOT NULL,
    status text NOT NULL,
    attempt_number integer NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NULL,
    error text NOT NULL DEFAULT ''
);

CREATE INDEX event_delivery_attempts_event_idx ON event_delivery_attempts (event_id, started_at DESC);
CREATE INDEX event_delivery_attempts_consumer_status_idx ON event_delivery_attempts (consumer_key, status, started_at DESC);

CREATE TABLE cronjob_definitions (
    key text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    schedule_kind text NOT NULL,
    schedule_expression text NOT NULL DEFAULT '',
    timezone text NOT NULL DEFAULT 'UTC',
    enabled boolean NOT NULL DEFAULT true,
    concurrency_policy text NOT NULL DEFAULT 'forbid',
    next_run_at timestamptz NULL,
    last_run_at timestamptz NULL,
    last_status text NOT NULL DEFAULT '',
    locked_by text NOT NULL DEFAULT '',
    locked_until timestamptz NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX cronjob_definitions_due_idx ON cronjob_definitions (enabled, next_run_at, key);
CREATE INDEX cronjob_definitions_locked_until_idx ON cronjob_definitions (locked_until);

CREATE TABLE cronjob_runs (
    id uuid PRIMARY KEY,
    job_key text NOT NULL,
    status text NOT NULL,
    scheduled_for timestamptz NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NULL,
    duration_ms bigint NOT NULL DEFAULT 0,
    trigger_type text NOT NULL,
    triggered_by_user_id uuid NULL,
    worker_id text NOT NULL,
    attempt_number integer NOT NULL DEFAULT 1,
    processed_count bigint NOT NULL DEFAULT 0,
    changed_count bigint NOT NULL DEFAULT 0,
    skipped_count bigint NOT NULL DEFAULT 0,
    metadata_json jsonb NOT NULL DEFAULT '{}',
    error text NOT NULL DEFAULT '',
    request_id text NOT NULL DEFAULT '',
    correlation_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX cronjob_runs_job_started_idx ON cronjob_runs (job_key, started_at DESC);
CREATE INDEX cronjob_runs_status_started_idx ON cronjob_runs (status, started_at DESC);
CREATE INDEX cronjob_runs_correlation_idx ON cronjob_runs (correlation_id);
