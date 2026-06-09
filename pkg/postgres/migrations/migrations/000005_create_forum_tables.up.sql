CREATE TABLE forum_categories (
    id uuid PRIMARY KEY,
    key text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    display_order integer NOT NULL DEFAULT 0,
    status text NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX forum_categories_key_active_idx ON forum_categories (key) WHERE deleted_at IS NULL;
CREATE INDEX forum_categories_order_active_idx ON forum_categories (display_order, id) WHERE deleted_at IS NULL;
CREATE INDEX forum_categories_status_active_idx ON forum_categories (status) WHERE deleted_at IS NULL;
CREATE INDEX forum_categories_deleted_at_idx ON forum_categories (deleted_at);

CREATE TABLE forums (
    id uuid PRIMARY KEY,
    category_id uuid NOT NULL,
    parent_forum_id uuid NULL,
    kind text NOT NULL,
    key text NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    display_order integer NOT NULL DEFAULT 0,
    path text NOT NULL,
    depth integer NOT NULL DEFAULT 0,
    external_url text NOT NULL DEFAULT '',
    icon_asset_id uuid NULL,
    thread_visibility_mode text NOT NULL,
    max_sticky_threads integer NOT NULL DEFAULT 0,
    default_thread_status text NOT NULL,
    status text NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX forums_key_active_idx ON forums (key) WHERE deleted_at IS NULL;
CREATE INDEX forums_category_parent_order_active_idx ON forums (category_id, parent_forum_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX forums_parent_order_active_idx ON forums (parent_forum_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX forums_path_active_idx ON forums (path) WHERE deleted_at IS NULL;
CREATE INDEX forums_slug_active_idx ON forums (slug) WHERE deleted_at IS NULL;
CREATE INDEX forums_status_active_idx ON forums (status) WHERE deleted_at IS NULL;
CREATE INDEX forums_deleted_at_idx ON forums (deleted_at);

CREATE TABLE forum_stats (
    forum_id uuid PRIMARY KEY,
    thread_count bigint NOT NULL DEFAULT 0,
    visible_thread_count bigint NOT NULL DEFAULT 0,
    post_count bigint NOT NULL DEFAULT 0,
    visible_post_count bigint NOT NULL DEFAULT 0,
    latest_thread_id uuid NULL,
    latest_post_id uuid NULL,
    latest_post_author_user_id uuid NULL,
    latest_post_at timestamptz NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX forum_stats_latest_post_at_idx ON forum_stats (latest_post_at);
