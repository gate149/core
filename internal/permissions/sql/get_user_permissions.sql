SELECT relation
FROM permissions
WHERE resource_type = $1
    AND resource_id = $2
    AND user_id = $3