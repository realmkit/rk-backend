CREATE TABLE groups (
    id uuid PRIMARY KEY,
    key text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    color text NOT NULL,
    weight integer NOT NULL DEFAULT 0,
    status text NOT NULL,
    icon_asset_id uuid NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX groups_key_active_idx ON groups (key) WHERE deleted_at IS NULL;
CREATE INDEX groups_status_active_idx ON groups (status) WHERE deleted_at IS NULL;
CREATE INDEX groups_weight_active_idx ON groups (weight) WHERE deleted_at IS NULL;
CREATE INDEX groups_deleted_at_idx ON groups (deleted_at);

CREATE TABLE group_memberships (
    id uuid PRIMARY KEY,
    group_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status text NOT NULL,
    assigned_by_user_id uuid NULL,
    assigned_reason text NOT NULL DEFAULT '',
    starts_at timestamptz NULL,
    expires_at timestamptz NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX group_memberships_group_user_active_idx ON group_memberships (group_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX group_memberships_user_active_idx ON group_memberships (user_id) WHERE deleted_at IS NULL;
CREATE INDEX group_memberships_group_active_idx ON group_memberships (group_id) WHERE deleted_at IS NULL;
CREATE INDEX group_memberships_status_active_idx ON group_memberships (status) WHERE deleted_at IS NULL;
CREATE INDEX group_memberships_expires_at_idx ON group_memberships (expires_at);
CREATE INDEX group_memberships_deleted_at_idx ON group_memberships (deleted_at);

CREATE TABLE permission_actions (
    id uuid PRIMARY KEY,
    action text NOT NULL,
    area text NOT NULL,
    scope_type text NOT NULL,
    label text NOT NULL,
    description text NOT NULL DEFAULT '',
    warning_level text NOT NULL DEFAULT 'normal',
    enabled boolean NOT NULL DEFAULT true,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX permission_actions_action_active_idx ON permission_actions (action) WHERE deleted_at IS NULL;
CREATE INDEX permission_actions_area_active_idx ON permission_actions (area) WHERE deleted_at IS NULL;
CREATE INDEX permission_actions_scope_active_idx ON permission_actions (scope_type) WHERE deleted_at IS NULL;
CREATE INDEX permission_actions_enabled_active_idx ON permission_actions (enabled) WHERE deleted_at IS NULL;
CREATE INDEX permission_actions_deleted_at_idx ON permission_actions (deleted_at);

CREATE TABLE permission_grants (
    id uuid PRIMARY KEY,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    action text NOT NULL,
    scope_type text NOT NULL,
    scope_id uuid NOT NULL,
    inherit boolean NOT NULL DEFAULT false,
    condition_key text NOT NULL DEFAULT '',
    created_by_user_id uuid NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX permission_grants_unique_active_idx ON permission_grants (subject_type, subject_id, action, scope_type, scope_id, inherit, condition_key) WHERE deleted_at IS NULL;
CREATE INDEX permission_grants_scope_action_idx ON permission_grants (scope_type, scope_id, action) WHERE deleted_at IS NULL;
CREATE INDEX permission_grants_subject_idx ON permission_grants (subject_type, subject_id) WHERE deleted_at IS NULL;
CREATE INDEX permission_grants_action_idx ON permission_grants (action) WHERE deleted_at IS NULL;
CREATE INDEX permission_grants_deleted_at_idx ON permission_grants (deleted_at);
