package service

import (
	"context"

	"github.com/MatTwix/Pull-Request-Assigner/internal/metrics"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo"
)

type UserService struct {
	userRepo repo.User
}

func NewUserService(userRepo repo.User) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*UserSetIsActiveOutput, error) {
	user, alreadyUpdated, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}

	outputUser := UserSetIsActiveOutputUser{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}

	output := UserSetIsActiveOutput{User: outputUser}

	if !alreadyUpdated {
		metrics.UserStatusChanges.WithLabelValues("setIsActive").Inc()
	}

	return &output, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) (*UserGetReviewOutput, error) {
	repositories, err := s.userRepo.GetReviewPRsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	output := UserGetReviewOutput{UserID: userID}
	for _, repository := range repositories {
		output.PullRequests = append(output.PullRequests, UserReviewOutputPR{
			PullRequestID:   repository.PullRequestID,
			PullRequestName: repository.PullRequestName,
			AuthorID:        repository.AuthorID,
			Status:          repository.Status,
		})
	}

	return &output, nil
}
