SELECT COUNT(*)
FROM problems
WHERE (
        $1::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM permissions perm
            WHERE perm.resource_id = problems.id
                AND perm.resource_type = 'problem'
                AND perm.user_id = $1
                AND perm.relation = 'owner'
        )
    )
    AND (
        $2::text IS NULL
        OR $2 = ''
        OR word_similarity(problems.title, $2) > 0.3
    )

