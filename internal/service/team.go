package service

import (
	"context"

	"github.com/MatTwix/Pull-Request-Assigner/internal/metrics"
	"github.com/MatTwix/Pull-Request-Assigner/internal/models"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo"
)

type TeamService struct {
	teamRepo repo.Team
}

func NewTeamService(teamRepo repo.Team) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

func (s *TeamService) AddTeam(ctx context.Context, input TeamAddInput) (*TeamAddOutput, error) {
	team := models.Team{
		TeamName: input.TeamName,
	}

	for _, member := range input.Members {
		team.Members = append(team.Members, models.User{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	createdTeam, err := s.teamRepo.CreateTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	outputTeam := TeamAddOutputTeam{TeamName: createdTeam.TeamName}

	for _, member := range createdTeam.Members {
		outputTeam.Members = append(outputTeam.Members, TeamOutputMember{
			UserID:   member.UserID,
			UserName: member.Username,
			IsActive: member.IsActive,
		})
	}

	output := TeamAddOutput{Team: outputTeam}

	metrics.TeamsCreated.Inc()
	metrics.UsersCreated.Add(float64(len(createdTeam.Members)))

	return &output, nil
}

func (s *TeamService) GetTeamByName(ctx context.Context, name string) (*TeamGetOutput, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, name)
	if err != nil {
		return nil, err
	}

	output := TeamGetOutput{TeamName: team.TeamName}

	for _, member := range team.Members {
		output.Members = append(output.Members, TeamOutputMember{
			UserID:   member.UserID,
			UserName: member.Username,
			IsActive: member.IsActive,
		})
	}

	return &output, nil
}
