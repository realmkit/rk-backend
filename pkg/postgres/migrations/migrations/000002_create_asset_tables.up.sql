CREATE TABLE assets (
    id uuid PRIMARY KEY,
    namespace text NOT NULL,
    path text NOT NULL DEFAULT '',
    filename text NOT NULL,
    display_name text NOT NULL,
    visibility text NOT NULL,
    status text NOT NULL,
    storage_key text NOT NULL,
    bucket text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL,
    etag text NOT NULL DEFAULT '',
    created_by_user_id uuid NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX assets_storage_key_active_idx ON assets (storage_key) WHERE deleted_at IS NULL;
CREATE INDEX assets_namespace_path_active_idx ON assets (namespace, path) WHERE deleted_at IS NULL;
CREATE INDEX assets_status_active_idx ON assets (status) WHERE deleted_at IS NULL;
CREATE INDEX assets_deleted_at_idx ON assets (deleted_at);
