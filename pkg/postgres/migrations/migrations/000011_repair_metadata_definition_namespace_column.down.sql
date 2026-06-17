ALTER TABLE metadata_metafield_definitions
    ADD COLUMN IF NOT EXISTS namespace varchar(64) NOT NULL DEFAULT 'default';
