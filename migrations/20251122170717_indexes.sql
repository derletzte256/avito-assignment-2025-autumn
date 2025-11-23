-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_user_team_name_is_active
    ON "user" (team_name, is_active);

CREATE INDEX idx_reviewer_user_id_pull_request_id
    ON reviewer (user_id, pull_request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reviewer_user_id_pull_request_id;
DROP INDEX IF EXISTS idx_user_team_name_is_active;
-- +goose StatementEnd
