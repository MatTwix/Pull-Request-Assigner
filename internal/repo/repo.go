package repo

import (
	"context"

	"github.com/MatTwix/Pull-Request-Assigner/internal/models"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/pgdb"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/database/postgres"
)

type User interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (userRes *models.User, alreadyUpdated bool, err error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error)
	GetReviewPRsByUserID(ctx context.Context, userID string) ([]models.PullRequest, error)
}

type PullRequest interface {
	CreatePR(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (pr *models.PullRequest, alreadyMerged bool, err error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (pullRequest *models.PullRequest, replacedBy string, err error)
}

type Team interface {
	CreateTeam(ctx context.Context, team models.Team) (*models.Team, error)
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	SetIsActiveTeam(ctx context.Context, teamName string, isActive bool) (int64, error)
}

type Repositories struct {
	User
	PullRequest
	Team
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		User:        pgdb.NewUserRepo(pg),
		PullRequest: pgdb.NewPullrequestRepo(pg),
		Team:        pgdb.NewTeamRepo(pg),
	}
}
