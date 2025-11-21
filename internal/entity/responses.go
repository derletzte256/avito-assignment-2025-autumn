package entity

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"

	ErrorCodeInternal     ErrorCode = "INTERNAL"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Info    string    `json:"info,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type CreateTeamResponse struct {
	Team *Team `json:"team"`
}

type PullRequestResponse struct {
	PR *PullRequest `json:"pr"`
}

type CreatePullRequestResponse = PullRequestResponse

type MergePullRequestResponse = PullRequestResponse

type SetUserActiveResponse struct {
	User *User `json:"user"`
}

type UserReviewListResponse struct {
	UserID       string         `json:"user_id"`
	PullRequests []*PullRequest `json:"pull_requests"`
}

type ReassignPullRequestResponse struct {
	PullRequestResponse
	ReplacedBy string `json:"replaced_by"`
}
