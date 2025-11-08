SELECT COUNT(*)
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