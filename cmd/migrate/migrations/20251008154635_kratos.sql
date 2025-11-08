-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS kratos_id varchar(255) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_users_kratos_id ON users(kratos_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_kratos_id;
ALTER TABLE users DROP COLUMN IF EXISTS kratos_id;
-- +goose StatementEnd
