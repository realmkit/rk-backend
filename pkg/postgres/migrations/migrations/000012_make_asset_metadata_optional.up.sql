UPDATE metadata_metafield_definitions
SET is_required = FALSE
WHERE owner_type = 'asset'
  AND is_required = TRUE;
