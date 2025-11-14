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

type TeamRepo struct {
	*postgres.Postgres
}

func NewTeamRepo(pg *postgres.Postgres) *TeamRepo {
	return &TeamRepo{pg}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team models.Team) (*models.Team, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, _ := r.Builder.
		Insert("teams").
		Columns("team_name").
		Values(team.TeamName).
		Suffix("RETURNING id").
		ToSql()

	if err := tx.QueryRow(ctx, sql, args...).Scan(&team.ID); err != nil {
		return nil, fmt.Errorf("failed to insert team: %w", err)
	}

	insert := r.Builder.
		Insert("users").
		Columns("user_id, username, team_name, is_active")

	for _, teamMember := range team.Members {
		insert = insert.Values(teamMember.UserID, teamMember.Username, team.TeamName, teamMember.IsActive)
	}

	sql, args, _ = insert.Suffix(`
		ON CONFLICT (user_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`).ToSql()

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to insert team member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &team, nil
}

func (r *TeamRepo) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	var teamID int
	sql, args, _ := r.Builder.
		Select("id").
		From("teams").
		Where("team_name = ?", name).
		Limit(1).
		ToSql()

	if err := r.Pool.QueryRow(ctx, sql, args...).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to check team: %w", err)
	}

	sql, args, _ = r.Builder.
		Select("id, user_id, username, is_active").
		From("users").
		Where("team_name = ?", name).
		OrderBy("id").
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	var teamMembers []models.User

	for rows.Next() {
		var teamMebmer models.User
		if err := rows.Scan(
			&teamMebmer.ID,
			&teamMebmer.UserID,
			&teamMebmer.Username,
			&teamMebmer.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		teamMembers = append(teamMembers, teamMebmer)
	}

	if len(teamMembers) == 0 {
		return nil, repoerrs.ErrNotFound
	}

	team := models.Team{
		ID:       teamID,
		TeamName: name,
		Members:  teamMembers,
	}

	return &team, nil
}
