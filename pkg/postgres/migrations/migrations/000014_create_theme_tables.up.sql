CREATE TABLE themes (
    id uuid PRIMARY KEY,
    key text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    status text NOT NULL,
    created_by_user_id uuid NULL,
    updated_by_user_id uuid NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX themes_key_active_idx ON themes (key) WHERE deleted_at IS NULL;
CREATE INDEX themes_status_active_idx ON themes (status, name) WHERE deleted_at IS NULL;
CREATE INDEX themes_deleted_at_idx ON themes (deleted_at);

CREATE TABLE theme_versions (
    id uuid PRIMARY KEY,
    theme_id uuid NOT NULL REFERENCES themes(id),
    semver text NOT NULL DEFAULT '',
    label text NOT NULL DEFAULT '',
    status text NOT NULL,
    source_kind text NOT NULL,
    source_reference text NOT NULL DEFAULT '',
    package_storage_key text NOT NULL DEFAULT '',
    package_size_bytes bigint NOT NULL DEFAULT 0,
    manifest_json jsonb NOT NULL DEFAULT '{}',
    settings_schema_json jsonb NOT NULL DEFAULT '{}',
    settings_data_json jsonb NOT NULL DEFAULT '{}',
    integrity_sha256 text NOT NULL DEFAULT '',
    published_at timestamptz NULL,
    published_by_user_id uuid NULL,
    created_by_user_id uuid NULL,
    updated_by_user_id uuid NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_versions_semver_active_idx ON theme_versions (theme_id, semver) WHERE semver <> '' AND deleted_at IS NULL;
CREATE INDEX theme_versions_theme_status_idx ON theme_versions (theme_id, status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX theme_versions_integrity_idx ON theme_versions (integrity_sha256) WHERE deleted_at IS NULL;
CREATE INDEX theme_versions_deleted_at_idx ON theme_versions (deleted_at);

CREATE TABLE theme_files (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    kind text NOT NULL,
    path text NOT NULL,
    content_sha256 text NOT NULL,
    content_storage_key text NOT NULL DEFAULT '',
    content_text text NOT NULL DEFAULT '',
    size_bytes bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_files_version_path_active_idx ON theme_files (version_id, path) WHERE deleted_at IS NULL;
CREATE INDEX theme_files_version_kind_idx ON theme_files (version_id, kind, path) WHERE deleted_at IS NULL;
CREATE INDEX theme_files_deleted_at_idx ON theme_files (deleted_at);

CREATE TABLE theme_assets (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    file_id uuid NOT NULL REFERENCES theme_files(id),
    path text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL DEFAULT 0,
    content_sha256 text NOT NULL,
    storage_key text NOT NULL,
    public_url text NOT NULL DEFAULT '',
    integrity_value text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_assets_version_path_active_idx ON theme_assets (version_id, path) WHERE deleted_at IS NULL;
CREATE INDEX theme_assets_storage_key_idx ON theme_assets (storage_key) WHERE deleted_at IS NULL;
CREATE INDEX theme_assets_deleted_at_idx ON theme_assets (deleted_at);

CREATE TABLE theme_templates (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    route_kind text NOT NULL,
    path text NOT NULL,
    layout_path text NOT NULL DEFAULT '',
    settings_json jsonb NOT NULL DEFAULT '{}',
    enabled boolean NOT NULL DEFAULT true,
    display_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_templates_route_active_idx ON theme_templates (version_id, route_kind) WHERE deleted_at IS NULL;
CREATE INDEX theme_templates_order_active_idx ON theme_templates (version_id, display_order, id) WHERE deleted_at IS NULL;

CREATE TABLE theme_sections (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    key text NOT NULL,
    name text NOT NULL DEFAULT '',
    path text NOT NULL,
    schema_json jsonb NOT NULL DEFAULT '{}',
    default_settings_json jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_sections_key_active_idx ON theme_sections (version_id, key) WHERE deleted_at IS NULL;

CREATE TABLE theme_snippets (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    key text NOT NULL,
    path text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_snippets_key_active_idx ON theme_snippets (version_id, key) WHERE deleted_at IS NULL;

CREATE TABLE theme_activations (
    id uuid PRIMARY KEY,
    theme_id uuid NOT NULL REFERENCES themes(id),
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    environment text NOT NULL,
    is_current boolean NOT NULL DEFAULT true,
    reason text NOT NULL DEFAULT '',
    settings_data_json jsonb NOT NULL DEFAULT '{}',
    activated_by_user_id uuid NULL,
    activated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_activations_current_environment_idx ON theme_activations (environment) WHERE is_current = true AND deleted_at IS NULL;
CREATE INDEX theme_activations_theme_environment_idx ON theme_activations (theme_id, environment, activated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX theme_activations_deleted_at_idx ON theme_activations (deleted_at);

CREATE TABLE theme_validation_issues (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    severity text NOT NULL,
    code text NOT NULL,
    path text NOT NULL DEFAULT '',
    message text NOT NULL,
    line integer NOT NULL DEFAULT 0,
    column_number integer NOT NULL DEFAULT 0,
    details_json jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX theme_validation_issues_version_severity_idx ON theme_validation_issues (version_id, severity, code) WHERE deleted_at IS NULL;
CREATE INDEX theme_validation_issues_deleted_at_idx ON theme_validation_issues (deleted_at);

CREATE TABLE theme_package_signatures (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    key_id text NOT NULL DEFAULT '',
    algorithm text NOT NULL,
    verification_status text NOT NULL,
    signature text NOT NULL DEFAULT '',
    signed_manifest_hash text NOT NULL DEFAULT '',
    verified_at timestamptz NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX theme_package_signatures_version_idx ON theme_package_signatures (version_id, verification_status) WHERE deleted_at IS NULL;
CREATE INDEX theme_package_signatures_key_idx ON theme_package_signatures (key_id) WHERE deleted_at IS NULL;

CREATE TABLE theme_signing_keys (
    id uuid PRIMARY KEY,
    key_id text NOT NULL,
    algorithm text NOT NULL,
    public_key text NOT NULL,
    trust_level text NOT NULL,
    status text NOT NULL,
    source text NOT NULL,
    not_before timestamptz NULL,
    not_after timestamptz NULL,
    created_by_user_id uuid NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz NULL,
    revoked_at timestamptz NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_signing_keys_key_id_active_idx ON theme_signing_keys (key_id) WHERE deleted_at IS NULL;
CREATE INDEX theme_signing_keys_status_idx ON theme_signing_keys (status, trust_level) WHERE deleted_at IS NULL;
CREATE INDEX theme_signing_keys_deleted_at_idx ON theme_signing_keys (deleted_at);

CREATE TABLE theme_preview_tokens (
    id uuid PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES theme_versions(id),
    token_hash text NOT NULL,
    persona_kind text NOT NULL,
    persona_source text NOT NULL,
    persona_user_id uuid NULL,
    expires_at timestamptz NOT NULL,
    created_by_user_id uuid NULL,
    created_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX theme_preview_tokens_hash_active_idx ON theme_preview_tokens (token_hash) WHERE deleted_at IS NULL;
CREATE INDEX theme_preview_tokens_version_idx ON theme_preview_tokens (version_id, expires_at) WHERE deleted_at IS NULL;
CREATE INDEX theme_preview_tokens_expiry_idx ON theme_preview_tokens (expires_at) WHERE deleted_at IS NULL;
