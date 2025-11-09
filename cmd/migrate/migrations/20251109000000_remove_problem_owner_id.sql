-- +goose Up
-- +goose StatementBegin
-- First, check if owner_id column exists in problems table
-- If it exists, migrate existing owner_id values to permissions table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'problems' AND column_name = 'owner_id'
    ) THEN
        -- Migrate existing owner_id values to permissions table
        INSERT INTO permissions (id, resource_type, resource_id, user_id, relation, created_at, updated_at)
        SELECT 
            gen_random_uuid(),
            'problem',
            p.id,
            p.owner_id,
            'owner',
            p.created_at,
            NOW()
        FROM problems p
        WHERE p.owner_id IS NOT NULL
        ON CONFLICT (resource_type, resource_id, user_id, relation) DO NOTHING;

        -- Now drop the owner_id column
        ALTER TABLE problems DROP COLUMN owner_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore owner_id column
ALTER TABLE problems ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Restore owner_id values from permissions table
UPDATE problems p
SET owner_id = perm.user_id
FROM permissions perm
WHERE perm.resource_type = 'problem'
  AND perm.resource_id = p.id
  AND perm.relation = 'owner'
  AND perm.user_id IS NOT NULL;

-- Note: We don't delete permissions here, just restore the column
-- +goose StatementEnd

