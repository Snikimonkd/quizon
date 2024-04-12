-- +goose Up
-- +goose StatementBegin
CREATE TABLE admins (
    name text PRIMARY KEY,
    password text NOT NULL
);

CREATE TABLE cookies (
    admin_name text NOT NULL,
    value uuid NOT NULL,
    expires timestamptz NOT NULL
);

CREATE UNIQUE INDEX cookies_uniq_ids ON cookies(admin_name, value);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cookies;
DROP TABLE users;
-- +goose StatementEnd
