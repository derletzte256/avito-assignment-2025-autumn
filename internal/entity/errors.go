package entity

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrPRMerged            = errors.New("pull request already merged")
	ErrNotAssignedReviewer = errors.New("user is not an assigned reviewer")
	ErrNoCandidate         = errors.New("no candidate reviewers available")
	ErrDuplicateUserIDs    = errors.New("duplicate user IDs provided")
	ErrUsersNotInSameTeam  = errors.New("users are not in the same team")
)
