-- +goose Up
-- +goose StatementBegin
CREATE TABLE games (
    id bigserial PRIMARY KEY,
    location text NOT NULL,
    start_time timestamptz NOT NULL,
    name text NOT NULL,
    teams_amount bigint NOT NULL,
    reserve bigint NOT NULL,
    registration_start timestamptz NOT NULL,
    comment text NOT NULL,
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE registrations (
    id bigserial PRIMARY KEY,
    game_id bigint NOT NULL REFERENCES games (id),
    team_id text NOT NULL,
    captain_name text NOT NULL,
    captain_group text NOT NULL,
    captain_telegram text NOT NULL,
    team_name text NOT NULL,
    team_size bigint NOT NULL,
    created_at timestamptz NOT NULL,
    deleted boolean NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE registrations;
DROP TABLE games;
-- +goose StatementEnd
