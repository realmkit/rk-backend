CREATE TABLE IF NOT EXISTS metadata_metafield_definitions (
    id uuid PRIMARY KEY,
    owner_type varchar(64) NOT NULL,
    namespace varchar(64) NOT NULL,
    key varchar(64) NOT NULL,
    name varchar(120) NOT NULL,
    description varchar(500),
    value_type varchar(64) NOT NULL,
    is_list boolean NOT NULL DEFAULT false,
    is_required boolean NOT NULL DEFAULT false,
    rules jsonb NOT NULL DEFAULT '{}',
    sort_order integer NOT NULL DEFAULT 0,
    active boolean NOT NULL DEFAULT true,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS metadata_metafield_values (
    id uuid PRIMARY KEY,
    definition_id uuid NOT NULL REFERENCES metadata_metafield_definitions(id),
    owner_type varchar(64) NOT NULL,
    owner_id uuid NOT NULL,
    value_json jsonb NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS metadata_metaobject_definitions (
    id uuid PRIMARY KEY,
    type varchar(64) NOT NULL,
    name varchar(120) NOT NULL,
    description varchar(500),
    field_definitions jsonb NOT NULL DEFAULT '[]',
    active boolean NOT NULL DEFAULT true,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS metadata_metaobject_entries (
    id uuid PRIMARY KEY,
    definition_id uuid NOT NULL REFERENCES metadata_metaobject_definitions(id),
    handle varchar(120) NOT NULL,
    display_name varchar(120) NOT NULL,
    field_values jsonb NOT NULL DEFAULT '{}',
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS metadata_metafield_definitions_owner_namespace_key_active_idx
ON metadata_metafield_definitions(owner_type, namespace, key)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS metadata_metafield_values_definition_owner_active_idx
ON metadata_metafield_values(definition_id, owner_type, owner_id)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS metadata_metaobject_definitions_type_active_idx
ON metadata_metaobject_definitions(type)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS metadata_metaobject_entries_definition_handle_active_idx
ON metadata_metaobject_entries(definition_id, handle)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS metadata_metafield_definitions_owner_active_sort_idx
ON metadata_metafield_definitions(owner_type, active, sort_order);

CREATE INDEX IF NOT EXISTS metadata_metafield_values_owner_idx
ON metadata_metafield_values(owner_type, owner_id);

CREATE INDEX IF NOT EXISTS metadata_metaobject_entries_definition_display_name_idx
ON metadata_metaobject_entries(definition_id, display_name);
