CREATE INDEX IF NOT EXISTS groups_search_active_idx
    ON groups
    USING gin (to_tsvector('simple', coalesce(key, '') || ' ' || coalesce(name, '') || ' ' || coalesce(description, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS assets_search_active_idx
    ON assets
    USING gin (to_tsvector('simple', coalesce(path, '') || ' ' || coalesce(filename, '') || ' ' || coalesce(display_name, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS user_provider_claim_cache_search_active_idx
    ON user_provider_claim_cache
    USING gin (to_tsvector('simple', coalesce(username, '') || ' ' || coalesce(email, '') || ' ' || coalesce(display_name, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS punishment_definitions_search_active_idx
    ON punishment_definitions
    USING gin (to_tsvector('simple', coalesce(key, '') || ' ' || coalesce(name, '') || ' ' || coalesce(description, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS punishments_search_active_idx
    ON punishments
    USING gin (to_tsvector('simple', coalesce(reason, '') || ' ' || coalesce(source, '') || ' ' || coalesce(issuer_key, '') || ' ' || coalesce(target_ip_hash, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ticket_definitions_search_active_idx
    ON ticket_definitions
    USING gin (to_tsvector('simple', coalesce(key, '') || ' ' || coalesce(name, '') || ' ' || coalesce(description, '')))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS tickets_search_active_idx
    ON tickets
    USING gin (to_tsvector('simple', coalesce(title, '') || ' ' || id::text || ' ' || coalesce(target_user_id::text, '') || ' ' || coalesce(punishment_id::text, '')))
    WHERE deleted_at IS NULL;
