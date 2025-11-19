package entity

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrPRMerged      = errors.New("pull request already merged")
	ErrNotAssigned   = errors.New("reviewer not assigned")
	ErrNoCandidate   = errors.New("no candidate available")
)
