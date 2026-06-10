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
    author_post_edit_window_seconds integer NOT NULL DEFAULT 600,
    author_post_delete_window_seconds integer NOT NULL DEFAULT 300,
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

CREATE TABLE forum_threads (
    id uuid PRIMARY KEY,
    forum_id uuid NOT NULL,
    author_user_id uuid NOT NULL,
    opener_post_id uuid NULL,
    latest_post_id uuid NULL,
    latest_post_author_user_id uuid NULL,
    latest_post_at timestamptz NULL,
    title text NOT NULL,
    slug text NOT NULL,
    status text NOT NULL,
    sticky_state text NOT NULL,
    sticky_order integer NOT NULL DEFAULT 0,
    sticky_until timestamptz NULL,
    locked_reason text NOT NULL DEFAULT '',
    reply_count bigint NOT NULL DEFAULT 0,
    visible_reply_count bigint NOT NULL DEFAULT 0,
    post_count bigint NOT NULL DEFAULT 0,
    visible_post_count bigint NOT NULL DEFAULT 0,
    like_count bigint NOT NULL DEFAULT 0,
    view_count bigint NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX forum_threads_forum_sticky_latest_active_idx ON forum_threads (forum_id, sticky_state, sticky_order, latest_post_at DESC, id) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_forum_latest_active_idx ON forum_threads (forum_id, latest_post_at DESC, id) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_forum_author_latest_active_idx ON forum_threads (forum_id, author_user_id, latest_post_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_forum_status_latest_active_idx ON forum_threads (forum_id, status, latest_post_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_slug_active_idx ON forum_threads (slug) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_like_latest_active_idx ON forum_threads (like_count DESC, latest_post_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_title_active_idx ON forum_threads (title) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_title_fts_idx ON forum_threads USING gin (to_tsvector('simple', title)) WHERE deleted_at IS NULL;
CREATE INDEX forum_threads_deleted_at_idx ON forum_threads (deleted_at);

CREATE TABLE forum_posts (
    id uuid PRIMARY KEY,
    thread_id uuid NOT NULL,
    forum_id uuid NOT NULL,
    author_user_id uuid NOT NULL,
    sequence bigint NOT NULL,
    status text NOT NULL,
    content_format text NOT NULL,
    content_document_json jsonb NOT NULL,
    content_text text NOT NULL,
    content_checksum text NOT NULL DEFAULT '',
    edited_at timestamptz NULL,
    edited_by_user_id uuid NULL,
    edit_count bigint NOT NULL DEFAULT 0,
    like_count bigint NOT NULL DEFAULT 0,
    reply_reference_count bigint NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX forum_posts_thread_sequence_active_idx ON forum_posts (thread_id, sequence) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_thread_status_sequence_active_idx ON forum_posts (thread_id, status, sequence) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_forum_created_active_idx ON forum_posts (forum_id, created_at DESC, id) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_forum_liked_active_idx ON forum_posts (forum_id, like_count DESC, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_author_created_active_idx ON forum_posts (author_user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_content_text_active_idx ON forum_posts (content_text) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_content_text_fts_idx ON forum_posts USING gin (to_tsvector('simple', content_text)) WHERE deleted_at IS NULL;
CREATE INDEX forum_posts_deleted_at_idx ON forum_posts (deleted_at);

CREATE TABLE forum_post_revisions (
    id uuid PRIMARY KEY,
    post_id uuid NOT NULL,
    edited_by_user_id uuid NOT NULL,
    previous_content_document_json jsonb NOT NULL,
    previous_content_text text NOT NULL,
    edit_reason text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX forum_post_revisions_post_created_idx ON forum_post_revisions (post_id, created_at DESC);

CREATE TABLE forum_post_references (
    id uuid PRIMARY KEY,
    source_post_id uuid NOT NULL,
    target_post_id uuid NULL,
    target_user_id uuid NULL,
    target_asset_id uuid NULL,
    reference_type text NOT NULL,
    quote_excerpt text NOT NULL DEFAULT '',
    link_url text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX forum_post_references_source_idx ON forum_post_references (source_post_id);
CREATE INDEX forum_post_references_target_post_idx ON forum_post_references (target_post_id);
CREATE INDEX forum_post_references_target_user_idx ON forum_post_references (target_user_id);
CREATE INDEX forum_post_references_target_asset_idx ON forum_post_references (target_asset_id);

CREATE TABLE forum_post_likes (
    id uuid PRIMARY KEY,
    post_id uuid NOT NULL,
    thread_id uuid NOT NULL,
    forum_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX forum_post_likes_post_user_active_idx ON forum_post_likes (post_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX forum_post_likes_user_created_active_idx ON forum_post_likes (user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_post_likes_forum_created_active_idx ON forum_post_likes (forum_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX forum_post_likes_thread_active_idx ON forum_post_likes (thread_id) WHERE deleted_at IS NULL;
CREATE INDEX forum_post_likes_deleted_at_idx ON forum_post_likes (deleted_at);

CREATE TABLE forum_thread_read_states (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    forum_id uuid NOT NULL,
    thread_id uuid NOT NULL,
    last_read_post_sequence bigint NOT NULL,
    last_read_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX forum_thread_read_states_user_thread_idx ON forum_thread_read_states (user_id, thread_id);
CREATE INDEX forum_thread_read_states_user_forum_read_idx ON forum_thread_read_states (user_id, forum_id, last_read_at DESC);
