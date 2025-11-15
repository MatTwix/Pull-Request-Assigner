package pgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/MatTwix/Pull-Request-Assigner/internal/models"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/repoerrs"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/database/postgres"
	"github.com/jackc/pgx/v5"
)

const MergedStatus = "MERGED"

type PullRequestRepo struct {
	*postgres.Postgres
}

func NewPullrequestRepo(pg *postgres.Postgres) *PullRequestRepo {
	return &PullRequestRepo{pg}
}

func (r *PullRequestRepo) CreatePR(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error) {
	sql, args, _ := r.Builder.
		Select("team_name").
		From("users").
		Where("user_id = ?", pr.AuthorID).
		ToSql()

	var teamName string

	if err := r.Pool.QueryRow(ctx, sql, args...).Scan(&teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get author team name: %w", err)
	}

	sql, args, _ = r.Builder.
		Select("user_id").
		From("users").
		Where("team_name = ? AND is_active = TRUE AND user_id != ?", teamName, pr.AuthorID).
		OrderBy("RANDOM()").
		Limit(2).
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		err := rows.Scan(&reviewerID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}

		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if len(pr.AssignedReviewers) == 0 {
		return nil, repoerrs.ErrNotFound
	}

	checkSQL, checkArgs, _ := r.Builder.
		Select("1").
		From("pull_requests").
		Where("pull_request_id = ?", pr.PullRequestID).
		ToSql()

	var exists int
	if err := r.Pool.QueryRow(ctx, checkSQL, checkArgs...).Scan(&exists); err == nil {
		return nil, repoerrs.ErrAlreadyExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check pr existence: %w", err)
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ = r.Builder.
		Insert("pull_requests").
		Columns("pull_request_id, pull_request_name, author_id").
		Values(
			pr.PullRequestID,
			pr.PullRequestName,
			pr.AuthorID,
		).
		Suffix("RETURNING id, status, created_at").
		ToSql()

	if err = tx.QueryRow(ctx, sql, args...).Scan(&pr.ID, &pr.Status, &pr.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to insert pr: %w", err)
	}

	insert := r.Builder.
		Insert("pull_request_reviewers").
		Columns("pull_request_id, reviewer_id")

	for _, reviewerID := range pr.AssignedReviewers {
		insert = insert.Values(pr.PullRequestID, reviewerID)
	}

	sql, args, _ = r.Builder.
		Update("users").
		Set("is_active", false).
		Where(squirrel.Eq{"user_id": pr.AssignedReviewers}).
		ToSql()

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to make reviewers not active: %w", err)
	}

	sql, args, _ = insert.ToSql()
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to insert reviewers: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepo) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var prevStatus string
	sql, args, _ := r.Builder.
		Select("status").
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": prID}).
		ToSql()

	if err := tx.QueryRow(ctx, sql, args...).Scan(&prevStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to check pr status: %w", err)
	}

	sql, args, _ = r.Builder.
		Update("pull_requests").
		Set("status", "MERGED").
		Set("merged_at", squirrel.Expr("COALESCE(merged_at, NOW())")).
		Where(squirrel.Eq{"pull_request_id": prID}).
		ToSql()

	cmdTag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec row: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, repoerrs.ErrNotFound
	}

	var reviewerIDs []string
	sql, args, _ = r.Builder.
		Select("reviewer_id").
		From("pull_request_reviewers").
		Where("pull_request_id = ?", prID).
		ToSql()

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select reviewers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string

		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer id: %w", err)
		}

		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	pr := models.PullRequest{
		PullRequestID:     prID,
		AssignedReviewers: reviewerIDs,
	}
	sql, args, _ = r.Builder.
		Select("id", "pull_request_name", "author_id", "status", "merged_at", "created_at").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&pr.ID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.MergedAt,
		&pr.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get updated pr: %w", err)
	}

	if len(reviewerIDs) > 0 && prevStatus != MergedStatus {
		sql, args, _ := r.Builder.
			Update("users").
			Set("is_active", true).
			Where(squirrel.Eq{"user_id": reviewerIDs}).
			ToSql()

		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return nil, fmt.Errorf("failed to make reviewers active: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepo) ReassignReviewer(ctx context.Context, prID, oldUserID string) (pullRequest *models.PullRequest, replacedBy string, err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var authorID, status string
	sql, args, _ := r.Builder.
		Select("author_id, status").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()

	if err := tx.QueryRow(ctx, sql, args...).Scan(&authorID, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", repoerrs.ErrNotFound
		}
		return nil, "", fmt.Errorf("failed to get pr: %w", err)
	}

	if status == MergedStatus {
		return nil, "", repoerrs.ErrReassignAfterMerge
	}

	sql, args, _ = r.Builder.
		Select("1").
		From("users").
		Where("user_id = ?", oldUserID).
		ToSql()

	var exists int
	if err := tx.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", repoerrs.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("failed to check old use existence: %w", err)
	}

	sql, args, _ = r.Builder.
		Select("1").
		From("pull_request_reviewers").
		Where("pull_request_id = ? AND reviewer_id = ?", prID, oldUserID).
		ToSql()

	if err := tx.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", repoerrs.ErrNotAssigned
		}
		return nil, "", fmt.Errorf("failed to check reviewer: %w", err)
	}

	var teamName string
	sql, args, _ = r.Builder.
		Select("team_name").
		From("users").
		Where("user_id = ?", oldUserID).
		ToSql()

	if err := tx.QueryRow(ctx, sql, args...).Scan(&teamName); err != nil {
		return nil, "", fmt.Errorf("failed to get team name: %w", err)
	}

	sql, args, _ = r.Builder.
		Select("user_id").
		From("users").
		Where(squirrel.Eq{"team_name": teamName, "is_active": true}).
		Where(squirrel.NotEq{"user_id": []string{authorID, oldUserID}}).
		OrderBy("RANDOM()").
		Limit(1).
		ToSql()

	var newReviewerID string
	if err := tx.QueryRow(ctx, sql, args...).Scan(&newReviewerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", repoerrs.ErrNoCandidate
		}
		return nil, "", fmt.Errorf("failed to find new reviewer: %w", err)
	}

	sql, args, _ = r.Builder.
		Update("pull_request_reviewers").
		Set("reviewer_id", newReviewerID).
		Where("pull_request_id = ? AND reviewer_id = ?", prID, oldUserID).
		ToSql()

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return nil, "", fmt.Errorf("failed to update reviewer: %w", err)
	}

	pr := models.PullRequest{
		PullRequestID: prID,
	}

	sql, args, _ = r.Builder.
		Select("id", "pull_request_name", "author_id", "status", "merged_at", "created_at").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&pr.ID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.MergedAt,
		&pr.CreatedAt,
	)

	if err != nil {
		return nil, "", fmt.Errorf("failed to get updated pr: %w", err)
	}

	var otherReviewerID string
	sql, args, _ = r.Builder.
		Select("reviewer_id").
		From("pull_request_reviewers").
		Where("pull_request_id = ? AND reviewer_id != ?", prID, newReviewerID).
		Limit(1).
		ToSql()

	err = tx.QueryRow(ctx, sql, args...).Scan(&otherReviewerID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, "", fmt.Errorf("failed to get reviewers list: %w", err)
	}

	pr.AssignedReviewers = append(pr.AssignedReviewers, newReviewerID)
	if otherReviewerID != "" {
		pr.AssignedReviewers = append(pr.AssignedReviewers, otherReviewerID)
	}

	sql, args, _ = r.Builder.
		Update("users").
		Set("is_active", true).
		Where("user_id = ?", oldUserID).
		ToSql()

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return nil, "", fmt.Errorf("failed to make old reviewer active: %w", err)
	}

	sql, args, _ = r.Builder.
		Update("users").
		Set("is_active", false).
		Where("user_id = ?", newReviewerID).
		ToSql()

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return nil, "", fmt.Errorf("failed to make new reviewer inactive: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", fmt.Errorf("failed to commit reassignment: %w", err)
	}

	return &pr, newReviewerID, nil
}
