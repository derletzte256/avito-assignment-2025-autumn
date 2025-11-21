package entity

type CreateTeamRequest = Team

type CreatePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
	Name          string `json:"pull_request_name"`
	AuthorID      string `json:"author_id"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}
