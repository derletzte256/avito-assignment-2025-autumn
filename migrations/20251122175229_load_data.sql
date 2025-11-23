-- +goose Up
-- +goose StatementBegin
TRUNCATE TABLE reviewer, pull_request, "user", team RESTART IDENTITY CASCADE;

COPY team (name, created_at)
FROM '/data/teams.csv'
WITH (FORMAT csv, HEADER true);

COPY "user" (id, username, is_active, team_name, created_at)
FROM '/data/users.csv'
WITH (FORMAT csv, HEADER true);

COPY pull_request (id, name, author_id, status_id, created_at, merged_at)
FROM '/data/pull_requests.csv'
WITH (FORMAT csv, HEADER true);

COPY reviewer (pull_request_id, user_id, created_at)
FROM '/data/reviews.csv'
WITH (FORMAT csv, HEADER true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove seeded sample data.
TRUNCATE TABLE reviewer, pull_request, "user", team RESTART IDENTITY CASCADE;
-- +goose StatementEnd
