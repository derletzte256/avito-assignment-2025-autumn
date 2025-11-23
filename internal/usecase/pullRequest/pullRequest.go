package pullRequest

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/logger"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/usecase"
	"go.uber.org/zap"
)

type UseCase struct {
	pullRequestRepo usecase.PullRequestRepository
	userRepo        usecase.UserRepository
	teamRepo        usecase.TeamRepository
	transactor      trm.Manager
}

func NewUseCase(pullRequestRepo usecase.PullRequestRepository, userRepo usecase.UserRepository, teamRepo usecase.TeamRepository, transactor trm.Manager) *UseCase {
	return &UseCase{
		pullRequestRepo: pullRequestRepo,
		teamRepo:        teamRepo,
		userRepo:        userRepo,
		transactor:      transactor,
	}
}

func (uc *UseCase) CreatePullRequest(ctx context.Context, pr *entity.CreatePullRequestRequest) (*entity.PullRequest, error) {
	l := logger.FromCtx(ctx)
	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		exists, err := uc.pullRequestRepo.CheckPullRequestIDExists(ctx, pr.PullRequestID)
		if err != nil {
			l.Warn("failed to check if PR exists", zap.Error(err))
			return err
		}
		if exists {
			return entity.ErrAlreadyExists
		}

		author, err := uc.userRepo.GetUserByID(ctx, pr.AuthorID)
		if err != nil {
			l.Warn("failed to get author by ID", zap.Error(err))
			return err
		}

		reviewers, err := uc.userRepo.GetReviewersForPullRequest(ctx, author.TeamName, author.ID)
		if err != nil {
			l.Warn("failed to get reviewers by ID", zap.Error(err))
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
			l.Warn("failed to create pull request", zap.Error(err))
			return err
		}

		if len(reviewersIDs) > 0 {
			if err = uc.pullRequestRepo.AssignReviewers(ctx, pr.PullRequestID, reviewersIDs); err != nil {
				l.Warn("failed to assign reviewers", zap.Error(err))
				return err
			}
		}

		return nil
	})

	if err != nil {
		l.Warn("failed to create pull request", zap.Error(err))
		return nil, err
	}

	createdPR, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get pull request", zap.Error(err))
		return nil, err
	}

	createdReviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get reviewers for pull request", zap.Error(err))
		return nil, err
	}
	createdPR.Reviewers = createdReviewers

	return createdPR, nil
}

func (uc *UseCase) MergePullRequest(ctx context.Context, pr *entity.MergePullRequestRequest) (*entity.PullRequest, error) {
	l := logger.FromCtx(ctx)
	exists, err := uc.pullRequestRepo.CheckPullRequestIDExists(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to check if PR exists", zap.Error(err))
		return nil, err
	}

	if !exists {
		return nil, entity.ErrNotFound
	}

	err = uc.pullRequestRepo.MergePullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to merge pull request", zap.Error(err))
		return nil, err
	}

	mergedPR, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get merged pull request", zap.Error(err))
		return nil, err
	}

	mergedReviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get reviewers for merged pull request", zap.Error(err))
		return nil, err
	}
	mergedPR.Reviewers = mergedReviewers

	return mergedPR, nil
}

func (uc *UseCase) ReassignPullRequest(ctx context.Context, pr *entity.ReassignPullRequestRequest) (*entity.ReassignPullRequestResponse, error) {
	l := logger.FromCtx(ctx)
	var replacedBy string
	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		pullRequest, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
		if err != nil {
			l.Warn("failed to get pull request", zap.Error(err))
			return err
		}

		if pullRequest.Status == entity.StatusMerged {
			return entity.ErrPRMerged
		}

		oldUser, err := uc.userRepo.GetUserByID(ctx, pr.OldUserID)
		if err != nil {
			l.Warn("failed to get old user", zap.Error(err))
			return err
		}

		reviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
		if err != nil {
			l.Warn("failed to get reviewers for pull request", zap.Error(err))
			return err
		}
		var isOldUserReviewer bool
		for _, reviewer := range reviewers {
			if reviewer == oldUser.ID {
				isOldUserReviewer = true
				break
			}
		}

		if !isOldUserReviewer {
			return entity.ErrNotAssignedReviewer
		}

		excludedIDs := make([]string, 0)
		excludedIDs = append(excludedIDs, reviewers...)
		excludedIDs = append(excludedIDs, pullRequest.AuthorID)

		newReviewer, err := uc.userRepo.GetReplacementReviewerForPullRequest(ctx, oldUser.TeamName, excludedIDs)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return entity.ErrNoCandidate
			}
			l.Warn("failed to get new reviewer", zap.Error(err))
			return err
		}

		replacedBy = newReviewer.ID

		err = uc.pullRequestRepo.RemoveReviewer(ctx, pr.PullRequestID, oldUser.ID)
		if err != nil {
			l.Warn("failed to remove old reviewer", zap.Error(err))
			return err
		}

		err = uc.pullRequestRepo.AddNewReviewer(ctx, pr.PullRequestID, newReviewer.ID)
		if err != nil {
			l.Warn("failed to assign new reviewer", zap.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	updatedPR, err := uc.pullRequestRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get updated pull request", zap.Error(err))
		return nil, err
	}

	updatedReviewers, err := uc.pullRequestRepo.GetReviewersByPullRequestID(ctx, pr.PullRequestID)
	if err != nil {
		l.Warn("failed to get reviewers for updated pull request", zap.Error(err))
		return nil, err
	}
	updatedPR.Reviewers = updatedReviewers

	pullRequestResponse := &entity.PullRequestResponse{
		PR: updatedPR,
	}

	result := &entity.ReassignPullRequestResponse{
		PullRequestResponse: pullRequestResponse,
		ReplacedBy:          replacedBy,
	}

	return result, nil
}
