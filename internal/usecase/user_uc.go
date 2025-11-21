package usecase

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.uber.org/zap"
)

type UserUseCase struct {
	userRepo        UserRepository
	pullRequestRepo PullRequestRepository
	transactor      trm.Manager
	logger          *zap.Logger
}

func NewUserUseCase(userRepo UserRepository, pullRequestRepo PullRequestRepository, transactor trm.Manager, logger *zap.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo:        userRepo,
		pullRequestRepo: pullRequestRepo,
		transactor:      transactor,
		logger:          logger,
	}
}

func (uc *UserUseCase) SetIsActive(ctx context.Context, req *entity.SetUserActiveRequest) (*entity.User, error) {
	exists, err := uc.userRepo.CheckUserExists(ctx, req.UserID)
	if err != nil {
		uc.logger.Warn("failed to check user existence", zap.Error(err))
		return nil, err
	}
	if !exists {
		return nil, entity.ErrNotFound
	}

	err = uc.userRepo.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		uc.logger.Warn("failed to set user active status", zap.Error(err))
		return nil, err
	}

	updatedUser, err := uc.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		uc.logger.Warn("failed to get updated user", zap.Error(err))
		return nil, err
	}

	return updatedUser, nil
}

func (uc *UserUseCase) GetReviewList(ctx context.Context, userID string) (*entity.UserReviewListResponse, error) {
	user, err := uc.userRepo.CheckUserExists(ctx, userID)
	if err != nil {
		uc.logger.Warn("failed to check user existence", zap.Error(err))
		return nil, err
	}

	if !user {
		return nil, entity.ErrNotFound
	}

	pullRequests, err := uc.pullRequestRepo.GetPullRequestsByReviewerID(ctx, userID)
	if err != nil {
		uc.logger.Warn("failed to get pull requests by reviewer ID", zap.Error(err))
		return nil, err
	}

	response := &entity.UserReviewListResponse{
		PullRequests: pullRequests,
		UserID:       userID,
	}

	return response, nil
}
