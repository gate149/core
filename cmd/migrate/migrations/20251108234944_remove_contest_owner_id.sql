-- +goose Up
-- +goose StatementBegin
-- First, migrate existing owner_id values to permissions table
INSERT INTO permissions (id, resource_type, resource_id, user_id, relation, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    'contest',
    c.id,
    c.owner_id,
    'owner',
    c.created_at,
    NOW()
FROM contests c
WHERE c.owner_id IS NOT NULL
ON CONFLICT (resource_type, resource_id, user_id, relation) DO NOTHING;

-- Now drop the owner_id column
ALTER TABLE contests DROP COLUMN owner_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore owner_id column
ALTER TABLE contests ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Restore owner_id values from permissions table
UPDATE contests c
SET owner_id = p.user_id
FROM permissions p
WHERE p.resource_type = 'contest'
  AND p.resource_id = c.id
  AND p.relation = 'owner'
  AND p.user_id IS NOT NULL;

-- Note: We don't delete permissions here, just restore the column
-- +goose StatementEnd

