package repoerrs

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	ErrUserNotFound       = errors.New("user not found")
	ErrReassignAfterMerge = errors.New("cannot reassign on merged PR")
	ErrNotAssigned        = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate        = errors.New("no active replacement candidate in team")
)
