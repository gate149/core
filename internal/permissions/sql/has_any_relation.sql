SELECT EXISTS(
        SELECT 1
        FROM permissions
        WHERE resource_type = $1
            AND resource_id = $2
            AND user_id = $3
            AND relation = ANY($4)
    )