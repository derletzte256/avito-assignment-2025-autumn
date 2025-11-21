package usecase

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.uber.org/zap"
)

type PullRequestUseCase struct {
	pullRequestRepo PullRequestRepository
	userRepo        UserRepository
	teamRepo        TeamRepository
	transactor      trm.Manager
	logger          *zap.Logger
}

func NewPullRequestUseCase(pullRequestRepo PullRequestRepository, userRepo UserRepository, teamRepo TeamRepository, transactor trm.Manager, logger *zap.Logger) *PullRequestUseCase {
	return &PullRequestUseCase{
		pullRequestRepo: pullRequestRepo,
		teamRepo:        teamRepo,
		userRepo:        userRepo,
		transactor:      transactor,
		logger:          logger,
	}
}

func (uc *PullRequestUseCase) CreatePullRequest(ctx context.Context, pr *entity.CreatePullRequestRequest) (*entity.PullRequest, error) {
	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		exists, err := uc.pullRequestRepo.CheckPullRequestIDExists(ctx, pr.PullRequestID)
		if err != nil {
			uc.logger.Warn("failed to check if PR exists", zap.Error(err))
			return err
		}
		if exists {
			return entity.ErrAlreadyExists
		}

		author, err := uc.userRepo.GetUserByID(ctx, pr.AuthorID)
		if err != nil {
			uc.logger.Warn("failed to get author by ID", zap.Error(err))
			return err
		}

		reviewers, err := uc.userRepo.GetReviewersForPullRequest(ctx, author.TeamName, author.ID)
		if err != nil {
			uc.logger.Warn("failed to get reviewers by ID", zap.Error(err))
			return err
		}

		reviewersIDs := make([]string, 0, len(reviewers))
		for _, reviewer := range reviewers {
			reviewersIDs = append(reviewersIDs, reviewer.ID)
		}

		newPR := &entity.PullRequest{
			ID:       pr.PullRequestID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   entity.StatusOpen,
		}

		if err = uc.pullRequestRepo.Create(ctx, newPR); err != nil {
			uc.logger.Warn("failed to create pull request", zap.Error(err))
			return err
		}

		if len(reviewersIDs) > 0 {
			if err = uc.pullRequestRepo.AssignReviewers(ctx, pr.PullRequestID, reviewersIDs); err != nil {
				uc.logger.Warn("failed to assign reviewers", zap.Error(err))
				return err
			}
		}

		return nil
	})

	if err != nil {
		uc.logger.Warn("failed to create pull request", zap.Error(err))
		return nil, err
	}

	createdPR, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to get pull request", zap.Error(err))
		return nil, err
	}

	createdReviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to get reviewers for pull request", zap.Error(err))
		return nil, err
	}
	createdPR.Reviewers = createdReviewers

	return createdPR, nil
}

func (uc *PullRequestUseCase) MergePullRequest(ctx context.Context, pr *entity.MergePullRequestRequest) (*entity.PullRequest, error) {
	exists, err := uc.pullRequestRepo.CheckPullRequestIDExists(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to check if PR exists", zap.Error(err))
		return nil, err
	}

	if !exists {
		return nil, entity.ErrNotFound
	}

	err = uc.pullRequestRepo.MergePullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to merge pull request", zap.Error(err))
		return nil, err
	}

	mergedPR, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to get merged pull request", zap.Error(err))
		return nil, err
	}

	mergedReviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
	if err != nil {
		uc.logger.Warn("failed to get reviewers for merged pull request", zap.Error(err))
		return nil, err
	}
	mergedPR.Reviewers = mergedReviewers

	return mergedPR, nil
}
