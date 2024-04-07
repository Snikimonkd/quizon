-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE index_counter;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE index_counter;
-- +goose StatementEnd
