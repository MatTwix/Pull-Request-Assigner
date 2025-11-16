package models

import "time"

type PullRequest struct {
	ID                 int        `db:"id"`
	PullRequestID      string     `db:"pull_request_id"`
	PullRequestName    string     `db:"pull_request_name"`
	AuthorID           string     `db:"author_id"`
	Status             string     `db:"status"`
	NeedsMoreReviewers bool       `db:"needs_more_reviewers"`
	CreatedAt          time.Time  `db:"created_at"`
	MergedAt           *time.Time `db:"merged_at"` // nullable

	AssignedReviewers []string `db:"-"` // reviewers uids
}
