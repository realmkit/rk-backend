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

CREATE TABLE authorization_relation_tuples (
    id uuid PRIMARY KEY,
    object_type text NOT NULL,
    object_id uuid NOT NULL,
    relation text NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid NOT NULL,
    subject_relation text NOT NULL DEFAULT '',
    created_by_user_id uuid NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX authorization_relation_tuples_unique_active_idx ON authorization_relation_tuples (object_type, object_id, relation, subject_type, subject_id, subject_relation) WHERE deleted_at IS NULL;
CREATE INDEX authorization_relation_tuples_object_idx ON authorization_relation_tuples (object_type, object_id, relation) WHERE deleted_at IS NULL;
CREATE INDEX authorization_relation_tuples_subject_idx ON authorization_relation_tuples (subject_type, subject_id) WHERE deleted_at IS NULL;
CREATE INDEX authorization_relation_tuples_subject_relation_idx ON authorization_relation_tuples (subject_type, subject_id, subject_relation) WHERE deleted_at IS NULL;
CREATE INDEX authorization_relation_tuples_deleted_at_idx ON authorization_relation_tuples (deleted_at);
