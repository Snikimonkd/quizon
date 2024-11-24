-- +goose Up
-- +goose StatementBegin
CREATE TABLE registrations (
    game_id bigint NOT NULL REFERENCES games(id),
    created_at timestamptz NOT NULL,

    -- название команды
    team_name text NOT NULL,
    -- имя капитана
    captain_name text NOT NULL,
    -- телефон капитана
    phone text NOT NULL,
    -- контакт для связи в тг
    telegram text NOT NULL,
    -- размер команды
    players_amount bigint NOT NULL,

    -- название и номер группы
    group_name text,
    -- айди команды
    team_id text
);

CREATE INDEX registrations_game_id ON registrations (game_id);
CREATE INDEX registrations_created_at ON registrations (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE registrations;
-- +goose StatementEnd
