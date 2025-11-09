UPDATE contests
SET title = COALESCE($2, title),
    is_private = COALESCE($3, is_private),
    monitor_enabled = COALESCE($4, monitor_enabled)
WHERE id = $1
