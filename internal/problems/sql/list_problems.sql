SELECT problems.id,
    problems.title,
    problems.memory_limit,
    problems.time_limit,
    problems.created_at,
    problems.updated_at
FROM problems
WHERE (
        (
            $1::uuid IS NULL
            AND problems.is_private = false
        )
        OR (
            $1::uuid IS NOT NULL
            AND EXISTS (
                SELECT 1
                FROM permissions perm
                WHERE perm.resource_id = problems.id
                    AND perm.resource_type = 'problem'
                    AND perm.user_id = $1
                    AND perm.relation = 'owner'
            )
        )
    )
    AND (
        $2::text IS NULL
        OR $2 = ''
        OR (
            CASE
                WHEN LENGTH($2) < 3 THEN problems.title ILIKE '%' || $2 || '%'
                ELSE word_similarity(problems.title, $2) > 0.3
            END
        )
    )
ORDER BY CASE
        WHEN $2::text IS NOT NULL
        AND $2 != ''
        AND LENGTH($2) >= 3 THEN word_similarity(problems.title, $2)
        ELSE NULL
    END DESC NULLS LAST,
    CASE
        WHEN $3::int < 0 THEN problems.created_at
        ELSE NULL
    END DESC,
    CASE
        WHEN $3::int >= 0 THEN problems.created_at
        ELSE NULL
    END ASC
LIMIT $4 OFFSET $5

