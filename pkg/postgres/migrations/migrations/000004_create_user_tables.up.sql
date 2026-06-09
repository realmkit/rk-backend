CREATE TABLE users (
    id uuid PRIMARY KEY,
    status text NOT NULL,
    avatar_asset_id uuid NULL,
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX users_status_active_idx ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX users_avatar_asset_id_active_idx ON users (avatar_asset_id) WHERE deleted_at IS NULL;
CREATE INDEX users_deleted_at_idx ON users (deleted_at);

CREATE TABLE user_identity_links (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    issuer text NOT NULL,
    subject text NOT NULL,
    subject_hash text NOT NULL,
    claims_hash text NOT NULL DEFAULT '',
    linked_at timestamptz NOT NULL,
    last_seen_at timestamptz NULL,
    last_synced_at timestamptz NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX user_identity_links_issuer_subject_active_idx ON user_identity_links (issuer, subject) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX user_identity_links_user_issuer_active_idx ON user_identity_links (user_id, issuer) WHERE deleted_at IS NULL;
CREATE INDEX user_identity_links_subject_hash_active_idx ON user_identity_links (subject_hash) WHERE deleted_at IS NULL;
CREATE INDEX user_identity_links_deleted_at_idx ON user_identity_links (deleted_at);

CREATE TABLE user_provider_claim_cache (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    issuer text NOT NULL,
    subject text NOT NULL,
    username text NOT NULL DEFAULT '',
    email text NOT NULL DEFAULT '',
    email_verified boolean NOT NULL DEFAULT false,
    display_name text NOT NULL DEFAULT '',
    picture_url text NOT NULL DEFAULT '',
    preferred_locale text NOT NULL DEFAULT '',
    claims_hash text NOT NULL DEFAULT '',
    synced_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX user_provider_claim_cache_user_active_idx ON user_provider_claim_cache (user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX user_provider_claim_cache_identity_active_idx ON user_provider_claim_cache (issuer, subject) WHERE deleted_at IS NULL;
CREATE INDEX user_provider_claim_cache_claims_hash_active_idx ON user_provider_claim_cache (claims_hash) WHERE deleted_at IS NULL;
CREATE INDEX user_provider_claim_cache_deleted_at_idx ON user_provider_claim_cache (deleted_at);
