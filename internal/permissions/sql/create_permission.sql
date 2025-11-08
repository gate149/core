INSERT INTO permissions (
        id,
        resource_type,
        resource_id,
        user_id,
        relation,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) ON CONFLICT (resource_type, resource_id, user_id, relation) DO NOTHING