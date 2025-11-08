SELECT c.id,
    c.title,
    c.is_private,
    c.monitor_enabled,
    c.created_at,
    c.updated_at
FROM contests c
WHERE (
        $1::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM permissions p
            WHERE p.resource_type = 'contest'
                AND p.resource_id = c.id
                AND p.user_id = $1
                AND p.relation = 'owner'
        )
    )
    AND (
        $2::text IS NULL
        OR $2 = ''
        OR word_similarity(c.title, $2) > 0.3
    )
ORDER BY CASE
        WHEN $2::text IS NOT NULL
        AND $2 != '' THEN word_similarity(c.title, $2)
        ELSE NULL
    END DESC NULLS LAST,
    CASE
        WHEN $3::bool = true THEN c.created_at
        ELSE NULL
    END DESC,
    CASE
        WHEN $3::bool = false THEN c.created_at
        ELSE NULL
    END ASC
LIMIT $4 OFFSET $5