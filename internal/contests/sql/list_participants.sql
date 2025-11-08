SELECT u.id,
    u.username,
    u.role,
    u.kratos_id,
    u.created_at,
    u.updated_at
FROM contest_user cu
    LEFT JOIN users u ON cu.user_id = u.id
WHERE contest_id = $1
LIMIT $2 OFFSET $3