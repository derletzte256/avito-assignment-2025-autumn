package entity

import "time"

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

type PullRequest struct {
	ID        string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    string     `json:"status"`
	Reviewers []string   `json:"assigned_reviewers,omitempty"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
}

type ReviewReassignment struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
	NewReviewerID string `json:"new_reviewer_id"`
}

type UnreassignedReview struct {
	PullRequestID string `json:"pull_request_id"`
	ReviewerID    string `json:"reviewer_id"`
}

type ReviewRecord struct {
	PullRequestID string
	ReviewerID    string
}
