-- tester/cmd/migrate/migrations/20251108120000_add_contests_owner.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE contests ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE contests DROP COLUMN owner_id;
-- +goose StatementEnd