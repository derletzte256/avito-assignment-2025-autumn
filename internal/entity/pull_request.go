package entity

import (
	"time"
)

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

type PullRequest struct {
	ID       string
	Name     string
	AuthorID string
	Status   string

	Reviewers []string
	CreatedAt time.Time
	MergedAt  *time.Time
}
