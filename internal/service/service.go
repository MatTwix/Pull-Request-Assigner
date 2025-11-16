package service

import (
	"context"
	"time"

	"github.com/MatTwix/Pull-Request-Assigner/internal/repo"
)

type Auth interface {
	IsUserKey(key string) bool
	IsAdminKey(key string) bool
}

type TeamAddInput struct {
	TeamName string
	Members  []TeamInputMember
}

type TeamInputMember struct {
	UserID   string
	Username string
	IsActive bool
}

type TeamAddOutput struct {
	Team TeamAddOutputTeam `json:"team"`
}

type TeamAddOutputTeam struct {
	TeamName string             `json:"team_name"`
	Members  []TeamOutputMember `json:"members"`
}

type TeamGetOutput struct {
	TeamName string             `json:"team_name"`
	Members  []TeamOutputMember `json:"members"`
}

type TeamOutputMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamSetIsActiveTeamOutput struct {
	UsersUpdated int64 `json:"users_updated"`
}

type Team interface {
	AddTeam(ctx context.Context, input TeamAddInput) (*TeamAddOutput, error)
	GetTeamByName(ctx context.Context, name string) (*TeamGetOutput, error)
	SetIsActiveTeam(ctx context.Context, teamName string, isActive bool) (*TeamSetIsActiveTeamOutput, error)
}

type UserSetIsActiveOutput struct {
	User UserSetIsActiveOutputUser `json:"user"`
}

type UserSetIsActiveOutputUser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserGetReviewOutput struct {
	UserID       string               `json:"user_id"`
	PullRequests []UserReviewOutputPR `json:"pull_requests"`
}

type UserReviewOutputPR struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type User interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*UserSetIsActiveOutput, error)
	GetReview(ctx context.Context, userID string) (*UserGetReviewOutput, error)
}

type PullRequestCreateInput struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
}

type PullRequestCreateOutput struct {
	PullRequest PullRequestCreateOutputPR `json:"pr"`
}

type PullRequestCreateOutputPR struct {
	PullRequestID     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt         time.Time `json:"created_at"`
}

type PullRequestMergeOutput struct {
	PullRequest PullRequestMergeOutputPR `json:"pr"`
}

type PullRequestMergeOutputPR struct {
	PullRequestID     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	MergedAt          time.Time `json:"mergedAt"`
}

type PullRequestReassignOutput struct {
	PullRequest PullRequestReassignOutputPR `json:"pr"`
	ReplacedBy  string                      `json:"replaced_by"`
}

type PullRequestReassignOutputPR struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type PullRequest interface {
	CreatePR(ctx context.Context, input PullRequestCreateInput) (*PullRequestCreateOutput, error)
	MergePR(ctx context.Context, prID string) (*PullRequestMergeOutput, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*PullRequestReassignOutput, error)
}

type Services struct {
	Auth        Auth
	Team        Team
	User        User
	PullRequest PullRequest
}

type ServicesDependencies struct {
	Repos *repo.Repositories

	AdminAPIKey string
	UserAPIKey  string
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		Auth:        NewAuthService(deps.UserAPIKey, deps.AdminAPIKey),
		User:        NewUserService(deps.Repos.User),
		Team:        NewTeamService(deps.Repos.Team),
		PullRequest: NewPullRequestService(deps.Repos.PullRequest),
	}
}
