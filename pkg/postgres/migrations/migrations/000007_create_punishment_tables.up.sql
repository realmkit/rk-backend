CREATE TABLE punishment_definitions (
    id uuid PRIMARY KEY,
    key text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    color text NOT NULL,
    severity integer NOT NULL DEFAULT 0,
    status text NOT NULL,
    default_duration_seconds bigint NULL,
    min_duration_seconds bigint NULL,
    max_duration_seconds bigint NULL,
    allow_permanent boolean NOT NULL DEFAULT false,
    requires_reason boolean NOT NULL DEFAULT true,
    requires_target_ip boolean NOT NULL DEFAULT false,
    display_order integer NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX punishment_definitions_key_active_idx ON punishment_definitions (key) WHERE deleted_at IS NULL;
CREATE INDEX punishment_definitions_status_order_active_idx ON punishment_definitions (status, severity, name) WHERE deleted_at IS NULL;
CREATE INDEX punishment_definitions_display_order_active_idx ON punishment_definitions (display_order, id) WHERE deleted_at IS NULL;
CREATE INDEX punishment_definitions_deleted_at_idx ON punishment_definitions (deleted_at);

CREATE TABLE punishment_definition_actions (
    id uuid PRIMARY KEY,
    definition_id uuid NOT NULL,
    target_system text NOT NULL,
    action_key text NOT NULL,
    effect text NOT NULL,
    configuration_json jsonb NOT NULL DEFAULT '{}',
    display_order integer NOT NULL DEFAULT 0,
    status text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX punishment_definition_actions_definition_order_idx ON punishment_definition_actions (definition_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX punishment_definition_actions_target_action_idx ON punishment_definition_actions (target_system, action_key) WHERE deleted_at IS NULL;
CREATE INDEX punishment_definition_actions_deleted_at_idx ON punishment_definition_actions (deleted_at);

CREATE TABLE punishments (
    id uuid PRIMARY KEY,
    definition_id uuid NOT NULL,
    target_user_id uuid NOT NULL,
    target_ip_hash text NOT NULL DEFAULT '',
    target_ip_ciphertext text NOT NULL DEFAULT '',
    issuer_type text NOT NULL,
    issuer_user_id uuid NULL,
    issuer_key text NOT NULL DEFAULT '',
    reason text NOT NULL,
    private_reason text NOT NULL DEFAULT '',
    status text NOT NULL,
    starts_at timestamptz NOT NULL,
    expires_at timestamptz NULL,
    revoked_at timestamptz NULL,
    revoked_by_user_id uuid NULL,
    revocation_reason text NOT NULL DEFAULT '',
    source text NOT NULL DEFAULT '',
    idempotency_key text NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX punishments_idempotency_key_idx ON punishments (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX punishments_target_status_expires_idx ON punishments (target_user_id, status, expires_at) WHERE deleted_at IS NULL;
CREATE INDEX punishments_target_ip_status_idx ON punishments (target_ip_hash, status, expires_at) WHERE deleted_at IS NULL;
CREATE INDEX punishments_definition_created_idx ON punishments (definition_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX punishments_issuer_created_idx ON punishments (issuer_user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX punishments_deleted_at_idx ON punishments (deleted_at);

CREATE TABLE punishment_action_snapshots (
    id uuid PRIMARY KEY,
    punishment_id uuid NOT NULL,
    definition_action_id uuid NOT NULL,
    target_system text NOT NULL,
    action_key text NOT NULL,
    effect text NOT NULL,
    configuration_json jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX punishment_action_snapshots_punishment_idx ON punishment_action_snapshots (punishment_id) WHERE deleted_at IS NULL;
CREATE INDEX punishment_action_snapshots_target_action_idx ON punishment_action_snapshots (target_system, action_key) WHERE deleted_at IS NULL;
CREATE INDEX punishment_action_snapshots_deleted_at_idx ON punishment_action_snapshots (deleted_at);

CREATE TABLE punishment_active_restrictions (
    id uuid PRIMARY KEY,
    punishment_id uuid NOT NULL,
    target_user_id uuid NOT NULL,
    action_key text NOT NULL,
    starts_at timestamptz NOT NULL,
    expires_at timestamptz NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX punishment_active_restrictions_user_action_idx ON punishment_active_restrictions (target_user_id, action_key, starts_at, expires_at) WHERE deleted_at IS NULL;
CREATE INDEX punishment_active_restrictions_punishment_idx ON punishment_active_restrictions (punishment_id) WHERE deleted_at IS NULL;
CREATE INDEX punishment_active_restrictions_deleted_at_idx ON punishment_active_restrictions (deleted_at);
