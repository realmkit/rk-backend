INSERT INTO groups(
    id,
    key,
    name,
    description,
    color,
    weight,
    status,
    icon_asset_id,
    version,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    '00000000-0000-0000-0000-000000000101',
    'administrator',
    'Administrator',
    'Highest rank of the server.',
    '#C74D53',
    10000,
    'system',
    NULL,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    NULL
)
ON CONFLICT(id) DO UPDATE SET
    key = 'administrator',
    name = 'Administrator',
    description = 'Highest rank of the server.',
    color = '#C74D53',
    weight = 10000,
    status = 'system',
    updated_at = CURRENT_TIMESTAMP,
    deleted_at = NULL;

INSERT INTO permission_grants(
    id,
    subject_type,
    subject_id,
    action,
    scope_type,
    scope_id,
    inherit,
    condition_key,
    created_by_user_id,
    created_at,
    deleted_at
) VALUES (
    '00000000-0000-0000-0000-000000000111',
    'group',
    '00000000-0000-0000-0000-000000000101',
    'groups.update',
    'group',
    '00000000-0000-0000-0000-000000000101',
    false,
    '',
    NULL,
    CURRENT_TIMESTAMP,
    NULL
)
ON CONFLICT(id) DO UPDATE SET
    deleted_at = NULL;
