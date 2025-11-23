package entity

type CreateTeamRequest = Team

type CreatePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required,min=1,max=64"`
	Name          string `json:"pull_request_name" validate:"required,min=1,max=128"`
	AuthorID      string `json:"author_id" validate:"required,min=1,max=64"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id" validate:"required,min=1,max=64"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required,min=1,max=64"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required,min=1,max=64"`
	OldUserID     string `json:"old_reviewer_id" validate:"required,min=1,max=64"`
}

type MassDeactivateUsersRequest struct {
	TeamName string   `json:"team_name" validate:"required,min=1,max=128"`
	UserIDs  []string `json:"user_ids" validate:"required,min=1,max=100,dive,required"`
}
