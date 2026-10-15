-- +goose Up
-- +goose StatementBegin
CREATE TABLE admins (
    login text NOT NULL,
    password text NOT NULL
);

CREATE UNIQUE INDEX admins_login_unique_idx ON admins(login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE admins;
-- +goose StatementEnd
