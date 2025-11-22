package user

import (
	"avito-assignment-2025-autumn/internal/entity"
	"avito-assignment-2025-autumn/internal/usecase"
	"context"
	"math/rand"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.uber.org/zap"
)

type UseCase struct {
	userRepo        usecase.UserRepository
	pullRequestRepo usecase.PullRequestRepository
	teamRepo        usecase.TeamRepository
	transactor      trm.Manager
	logger          *zap.Logger
}

func NewUseCase(userRepo usecase.UserRepository, pullRequestRepo usecase.PullRequestRepository, teamRepo usecase.TeamRepository, transactor trm.Manager, logger *zap.Logger) *UseCase {
	return &UseCase{
		userRepo:        userRepo,
		pullRequestRepo: pullRequestRepo,
		teamRepo:        teamRepo,
		transactor:      transactor,
		logger:          logger,
	}
}

func (uc *UseCase) SetIsActive(ctx context.Context, req *entity.SetUserActiveRequest) (*entity.User, error) {
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

func (uc *UseCase) GetReviewList(ctx context.Context, userID string) (*entity.UserReviewListResponse, error) {
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

func (uc *UseCase) GetStatistics(ctx context.Context) (*entity.StatsByUsersResponse, error) {
	usersIDs, err := uc.userRepo.GetAllUsersIDs(ctx)
	if err != nil {
		uc.logger.Warn("failed to get all user IDs", zap.Error(err))
		return nil, err
	}

	openMap, mergedMap, err := uc.pullRequestRepo.GetOpenAndMergedReviewStatisticsForUsers(ctx)
	if err != nil {
		uc.logger.Warn("failed to get review statistics for users", zap.Error(err))
		return nil, err
	}

	authorMap, err := uc.pullRequestRepo.GetAuthorStatisticsForUsers(ctx)
	if err != nil {
		uc.logger.Warn("failed to get author statistics for users", zap.Error(err))
		return nil, err
	}

	for _, userID := range usersIDs {
		if _, exists := openMap[userID]; !exists {
			openMap[userID] = 0
		}
		if _, exists := mergedMap[userID]; !exists {
			mergedMap[userID] = 0
		}
		if _, exists := authorMap[userID]; !exists {
			authorMap[userID] = 0
		}
	}

	statsMap := make(map[string]*entity.UserStatistics)
	for _, userID := range usersIDs {
		statsMap[userID] = &entity.UserStatistics{
			UserID:          userID,
			AuthoredPRCount: authorMap[userID],
			OnReviewPRCount: openMap[userID],
			ReviewedPRCount: mergedMap[userID],
		}
	}

	return &entity.StatsByUsersResponse{UsersStat: statsMap}, nil
}

func checkUsersSameTeam(users []*entity.User) bool {
	if len(users) == 0 {
		return true
	}
	teamName := users[0].TeamName
	for _, user := range users[1:] {
		if user.TeamName != teamName {
			return false
		}
	}
	return true
}

func checkUserIDsUnique(ids []string) bool {
	idsMap := make(map[string]struct{})
	for _, id := range ids {
		if id != "" {
			if _, exists := idsMap[id]; exists {
				return false
			}
			idsMap[id] = struct{}{}
		}
	}
	return true
}

func (uc *UseCase) MassDeactivateUsers(ctx context.Context, req *entity.MassDeactivateUsersRequest) (*entity.MassDeactivateUsersResponse, error) {
	if !checkUserIDsUnique(req.UserIDs) {
		return nil, entity.ErrDuplicateUserIDs
	}

	reviewReassignments := make([]*entity.ReviewReassignment, 0)
	unreassignedReviews := make([]*entity.UnreassignedReview, 0)

	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		if _, err := uc.validateTeamAndUsers(ctx, req.TeamName, req.UserIDs); err != nil {
			return err
		}

		candidateIDs, excludeSet, err := uc.getCandidateReviewerIDs(ctx, req.TeamName, req.UserIDs)
		if err != nil {
			return err
		}

		var oldReviewsToRemove []*entity.ReviewRecord
		var newReviewsToAdd []*entity.ReviewRecord
		reviewReassignments, unreassignedReviews, oldReviewsToRemove, newReviewsToAdd, err = uc.buildReviewReassignmentPlan(ctx, candidateIDs, excludeSet, req.UserIDs)
		if err != nil {
			return err
		}

		if err := uc.applyMassDeactivation(ctx, req.UserIDs, oldReviewsToRemove, newReviewsToAdd); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &entity.MassDeactivateUsersResponse{
		TeamName:            req.TeamName,
		ReviewReassignments: reviewReassignments,
		UnreassignedReviews: unreassignedReviews,
	}, nil
}

func (uc *UseCase) validateTeamAndUsers(ctx context.Context, teamName string, userIDs []string) ([]*entity.User, error) {
	exists, err := uc.teamRepo.CheckTeamNameExists(ctx, teamName)
	if err != nil {
		uc.logger.Warn("failed to check team existence", zap.Error(err))
		return nil, err
	}
	if !exists {
		uc.logger.Warn("team does not exist", zap.String("teamName", teamName))
		return nil, entity.ErrNotFound
	}

	users, err := uc.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		uc.logger.Warn("failed to get users by IDs", zap.Error(err))
		return nil, err
	}

	if len(users) == 0 {
		uc.logger.Warn("no users found for deactivation", zap.String("teamName", teamName))
		return nil, entity.ErrNotFound
	}

	if !checkUsersSameTeam(users) {
		uc.logger.Warn("users are not from the same team")
		return nil, entity.ErrUsersNotInSameTeam
	}

	if users[0].TeamName != teamName {
		uc.logger.Warn("users team does not match request team", zap.String("teamName", teamName), zap.String("usersTeam", users[0].TeamName))
		return nil, entity.ErrUsersNotInSameTeam
	}

	return users, nil
}

func (uc *UseCase) getCandidateReviewerIDs(ctx context.Context, teamName string, userIDs []string) ([]string, map[string]struct{}, error) {
	candidateIDs, err := uc.userRepo.GetActiveUsersIDsByTeamName(ctx, teamName, userIDs)
	if err != nil {
		uc.logger.Warn("failed to get active users by team name", zap.Error(err))
		return nil, nil, err
	}

	excludeSet := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		excludeSet[id] = struct{}{}
	}

	if len(candidateIDs) == 0 {
		return candidateIDs, excludeSet, nil
	}

	filteredCandidates := make([]string, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		if _, excluded := excludeSet[id]; !excluded {
			filteredCandidates = append(filteredCandidates, id)
		}
	}

	return filteredCandidates, excludeSet, nil
}

func (uc *UseCase) buildReviewReassignmentPlan(
	ctx context.Context,
	candidateIDs []string,
	excludeSet map[string]struct{},
	reviewerIDs []string,
) ([]*entity.ReviewReassignment, []*entity.UnreassignedReview, []*entity.ReviewRecord, []*entity.ReviewRecord, error) {
	reviewReassignments := make([]*entity.ReviewReassignment, 0)
	unreassignedReviews := make([]*entity.UnreassignedReview, 0)
	oldReviewsToRemove := make([]*entity.ReviewRecord, 0)
	newReviewsToAdd := make([]*entity.ReviewRecord, 0)

	reviewRecords, err := uc.pullRequestRepo.GetReviewsByReviewerIDs(ctx, reviewerIDs)
	if err != nil {
		uc.logger.Warn("failed to get reviews by reviewer IDs", zap.Error(err))
		return nil, nil, nil, nil, err
	}

	if len(reviewRecords) == 0 {
		return reviewReassignments, unreassignedReviews, oldReviewsToRemove, newReviewsToAdd, nil
	}

	if len(candidateIDs) == 0 {
		for _, record := range reviewRecords {
			if record == nil {
				continue
			}
			unreassignedReviews = append(unreassignedReviews, &entity.UnreassignedReview{
				PullRequestID: record.PullRequestID,
				ReviewerID:    record.ReviewerID,
			})
		}
		return reviewReassignments, unreassignedReviews, oldReviewsToRemove, newReviewsToAdd, nil
	}

	prIDSet := make(map[string]struct{})
	for _, record := range reviewRecords {
		if record != nil {
			prIDSet[record.PullRequestID] = struct{}{}
		}
	}

	prIDs := make([]string, 0, len(prIDSet))
	for id := range prIDSet {
		prIDs = append(prIDs, id)
	}

	reviewersMap, err := uc.pullRequestRepo.GetReviewersByPullRequestIDs(ctx, prIDs)
	if err != nil {
		uc.logger.Warn("failed to get reviewers by pull request IDs", zap.Error(err))
		return nil, nil, nil, nil, err
	}

	reviewersByPR := make(map[string]map[string]struct{}, len(reviewersMap))
	for prID, reviewers := range reviewersMap {
		set := make(map[string]struct{}, len(reviewers))
		for _, reviewerID := range reviewers {
			if reviewerID != "" {
				set[reviewerID] = struct{}{}
			}
		}
		reviewersByPR[prID] = set
	}

	currentIndex := 0
	if len(candidateIDs) > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		currentIndex = r.Intn(len(candidateIDs))
	}

	for _, record := range reviewRecords {
		if record == nil {
			continue
		}

		prID := record.PullRequestID
		oldReviewerID := record.ReviewerID

		existingReviewers, ok := reviewersByPR[prID]
		if !ok {
			existingReviewers = make(map[string]struct{})
			reviewersByPR[prID] = existingReviewers
		}

		var replacementID string
		for i := 0; i < len(candidateIDs); i++ {
			idx := (currentIndex + i) % len(candidateIDs)
			candidateID := candidateIDs[idx]
			if candidateID == "" {
				continue
			}
			if _, excluded := excludeSet[candidateID]; excluded {
				continue
			}
			if _, alreadyAssigned := existingReviewers[candidateID]; alreadyAssigned {
				continue
			}
			replacementID = candidateID
			currentIndex = (idx + 1) % len(candidateIDs)
			break
		}

		if replacementID == "" {
			unreassignedReviews = append(unreassignedReviews, &entity.UnreassignedReview{
				PullRequestID: prID,
				ReviewerID:    oldReviewerID,
			})
			continue
		}

		reviewReassignments = append(reviewReassignments, &entity.ReviewReassignment{
			PullRequestID: prID,
			OldReviewerID: oldReviewerID,
			NewReviewerID: replacementID,
		})

		oldReviewsToRemove = append(oldReviewsToRemove, &entity.ReviewRecord{
			PullRequestID: prID,
			ReviewerID:    oldReviewerID,
		})
		newReviewsToAdd = append(newReviewsToAdd, &entity.ReviewRecord{
			PullRequestID: prID,
			ReviewerID:    replacementID,
		})

		existingReviewers[replacementID] = struct{}{}
	}

	return reviewReassignments, unreassignedReviews, oldReviewsToRemove, newReviewsToAdd, nil
}

func (uc *UseCase) applyMassDeactivation(
	ctx context.Context,
	userIDs []string,
	oldReviewsToRemove []*entity.ReviewRecord,
	newReviewsToAdd []*entity.ReviewRecord,
) error {
	if err := uc.userRepo.DeactivateUsers(ctx, userIDs); err != nil {
		uc.logger.Warn("failed to deactivate users", zap.Error(err))
		return err
	}

	if len(oldReviewsToRemove) > 0 {
		if err := uc.pullRequestRepo.RemoveReviewersBatch(ctx, oldReviewsToRemove); err != nil {
			uc.logger.Warn("failed to remove old reviewers in batch", zap.Error(err))
			return err
		}
	}

	if len(newReviewsToAdd) > 0 {
		if err := uc.pullRequestRepo.AddReviewersBatch(ctx, newReviewsToAdd); err != nil {
			uc.logger.Warn("failed to add new reviewers in batch", zap.Error(err))
			return err
		}
	}

	return nil
}
