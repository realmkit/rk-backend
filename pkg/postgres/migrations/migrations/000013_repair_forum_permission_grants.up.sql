UPDATE forum_permission_grants
SET deleted_at = CURRENT_TIMESTAMP
WHERE deleted_at IS NULL
  AND (
    action = 'forums.manage_forum'
    OR (subject_type = 'public' AND action <> 'forums.view')
  );
