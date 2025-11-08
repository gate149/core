-- +goose Up
-- +goose StatementBegin
ALTER TABLE contests ADD COLUMN is_private BOOLEAN DEFAULT true;
ALTER TABLE contests ADD COLUMN monitor_enabled BOOLEAN DEFAULT false;
ALTER TABLE problems ADD COLUMN is_private BOOLEAN DEFAULT true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE contests DROP COLUMN monitor_enabled;
ALTER TABLE contests DROP COLUMN is_private;
ALTER TABLE problems DROP COLUMN is_private;
-- +goose StatementEnd


