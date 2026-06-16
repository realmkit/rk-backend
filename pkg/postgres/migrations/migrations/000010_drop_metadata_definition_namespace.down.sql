DROP INDEX IF EXISTS metadata_metafield_definitions_owner_key_active_idx;

ALTER TABLE metadata_metafield_definitions
    ADD COLUMN namespace varchar(64) NOT NULL DEFAULT 'default';

CREATE UNIQUE INDEX IF NOT EXISTS metadata_metafield_definitions_owner_namespace_key_active_idx
ON metadata_metafield_definitions(owner_type, namespace, key)
WHERE deleted_at IS NULL;
