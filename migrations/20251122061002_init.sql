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

CREATE TABLE pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL REFERENCES users(user_id),
    status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ
);

CREATE INDEX idx_pull_requests_author_id
    ON pull_requests(author_id);


CREATE TABLE pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer_id
    ON pull_request_reviewers(reviewer_id);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
DROP TABLE IF EXISTS pull_request_reviewers;

DROP INDEX IF EXISTS idx_pull_requests_author_id;
DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_users_team_name;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS teams;

-- +goose StatementEnd
