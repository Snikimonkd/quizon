-- +goose Up
-- +goose StatementBegin
CREATE TABLE games (
    id bigserial NOT NULL,
    created_at timestamptz NOT NULL,

    -- время начала игры
    start_time timestamptz NOT NULL,
    -- местро проведения игры
    location text NOT NULL,
    -- название игры
    name text NOT NULL,

    -- количество команд на игре
    main_amount bigint NOT NULL,
    -- количество команд в резерве
    reserve_amount bigint NOT NULL,

    -- время открытия регистрации
    registration_open_time timestamptz NOT NULL
);
CREATE UNIQUE INDEX games_id_idx ON games (id);
CREATE INDEX games_date_idx ON games (start_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE games;
-- +goose StatementEnd
