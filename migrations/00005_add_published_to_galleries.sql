-- +goose Up
-- +goose StatementBegin
ALTER TABLE galleries
  ADD COLUMN published BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE galleries
  DROP COLUMN published;
-- +goose StatementEnd
