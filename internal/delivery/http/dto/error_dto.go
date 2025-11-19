package dto

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrorCodeInternal    ErrorCode = "INTERNAL"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Info    string    `json:"info,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}
