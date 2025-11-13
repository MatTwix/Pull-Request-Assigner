package models

type PullRequestReviewer struct {
	PullRequestID string `db:"pull_request_id"`
	ReviewerID    string `db:"reviewer_id"`
}
