-- +goose Up
-- +goose StatementBegin
CREATE TABLE team (
	name TEXT PRIMARY KEY
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "user" (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	team_name TEXT NOT NULL REFERENCES team(name) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pull_request_status (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL CONSTRAINT pull_request_status_name_unique UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO pull_request_status (name) VALUES ('OPEN'), ('MERGED');

CREATE TABLE pull_request (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    status_id BIGINT NOT NULL REFERENCES pull_request_status(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
);

CREATE TABLE reviewer (
    pull_request_id TEXT NOT NULL REFERENCES pull_request(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reviewer;
DROP TABLE IF EXISTS pull_request;
DROP TABLE IF EXISTS pull_request_status;
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS team;
-- +goose StatementEnd
