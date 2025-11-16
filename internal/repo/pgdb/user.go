package pgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/MatTwix/Pull-Request-Assigner/internal/models"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/repoerrs"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/database/postgres"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (userRes *models.User, alreadyUpdated bool, err error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, false, err
	}

	alreadyUpdated = user.IsActive == isActive

	if !alreadyUpdated {
		sql, args, _ := r.Builder.
			Update("users").
			Set("is_active", isActive).
			Where("user_id = ?", userID).
			ToSql()

		if _, err = r.Pool.Exec(ctx, sql, args...); err != nil {
			return nil, false, fmt.Errorf("failed to execute sql request: %w", err)
		}

		user.IsActive = isActive
	}

	return user, alreadyUpdated, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	sql, args, _ := r.Builder.
		Select("id, username, team_name, is_active").
		From("users").
		Where("user_id = ?", userID).
		ToSql()

	user := models.User{
		UserID: userID,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	sql, args, _ := r.Builder.
		Select("id, user_id, username, is_active").
		From("users").Where("team_name = ? AND is_active = TRUE", teamName).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		user := models.User{
			TeamName: teamName,
		}
		err := rows.Scan(
			&user.ID,
			&user.UserID,
			&user.Username,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, repoerrs.ErrNotFound
	}

	return users, nil
}

func (r *UserRepo) GetReviewPRsByUserID(ctx context.Context, userID string) ([]models.PullRequest, error) {
	sql, args, _ := r.Builder.
		Select("pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status").
		From("pull_requests pr").
		Join("pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where("prr.reviewer_id = ?", userID).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	var pullRequests []models.PullRequest
	for rows.Next() {
		var pullRequest models.PullRequest

		err := rows.Scan(
			&pullRequest.PullRequestID,
			&pullRequest.PullRequestName,
			&pullRequest.AuthorID,
			&pullRequest.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		pullRequests = append(pullRequests, pullRequest)
	}

	return pullRequests, nil
}
