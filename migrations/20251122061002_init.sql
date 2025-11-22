-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE users (
    user_id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    team_name TEXT NULL REFERENCES teams(team_name),
    is_active BOOLEAN NOT NULL
);

CREATE INDEX idx_users_team_name
    ON users(team_name);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
