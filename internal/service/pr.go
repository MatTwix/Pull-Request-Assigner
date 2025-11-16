package service

import (
	"context"

	"github.com/MatTwix/Pull-Request-Assigner/internal/metrics"
	"github.com/MatTwix/Pull-Request-Assigner/internal/models"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo"
)

type PullRequestService struct {
	pullRequestRepo repo.PullRequest
}

func NewPullRequestService(pullRequestRepo repo.PullRequest) *PullRequestService {
	return &PullRequestService{pullRequestRepo: pullRequestRepo}
}

func (s *PullRequestService) CreatePR(ctx context.Context, input PullRequestCreateInput) (*PullRequestCreateOutput, error) {
	pullRequest := models.PullRequest{
		PullRequestID:   input.PullRequestID,
		PullRequestName: input.PullRequestName,
		AuthorID:        input.AuthorID,
	}

	createdPullRequest, err := s.pullRequestRepo.CreatePR(ctx, pullRequest)
	if err != nil {
		return nil, err
	}

	outputPR := PullRequestCreateOutputPR{
		PullRequestID:     createdPullRequest.PullRequestID,
		PullRequestName:   createdPullRequest.PullRequestName,
		AuthorID:          createdPullRequest.AuthorID,
		Status:            createdPullRequest.Status,
		AssignedReviewers: createdPullRequest.AssignedReviewers,
		CreatedAt:         createdPullRequest.CreatedAt,
	}

	output := PullRequestCreateOutput{PullRequest: outputPR}

	metrics.PRCreated.Inc()
	return &output, nil
}

func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*PullRequestMergeOutput, error) {
	pullRequest, alreadyMerged, err := s.pullRequestRepo.MergePR(ctx, prID)
	if err != nil {
		return nil, err
	}

	outputPR := PullRequestMergeOutputPR{
		PullRequestID:     pullRequest.PullRequestID,
		PullRequestName:   pullRequest.PullRequestName,
		AuthorID:          pullRequest.AuthorID,
		Status:            pullRequest.Status,
		AssignedReviewers: pullRequest.AssignedReviewers,
		MergedAt:          *pullRequest.MergedAt,
	}

	output := PullRequestMergeOutput{PullRequest: outputPR}

	if !alreadyMerged {
		metrics.PRMerged.Inc()
	}
	return &output, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*PullRequestReassignOutput, error) {
	pullRequest, replacedBy, err := s.pullRequestRepo.ReassignReviewer(ctx, prID, oldUserID)
	if err != nil {
		return nil, err
	}

	output := PullRequestReassignOutput{
		ReplacedBy: replacedBy,
		PullRequest: PullRequestReassignOutputPR{
			PullRequestID:     pullRequest.PullRequestID,
			PullRequestName:   pullRequest.PullRequestName,
			AuthorID:          pullRequest.AuthorID,
			Status:            pullRequest.Status,
			AssignedReviewers: pullRequest.AssignedReviewers,
		},
	}

	return &output, nil
}
