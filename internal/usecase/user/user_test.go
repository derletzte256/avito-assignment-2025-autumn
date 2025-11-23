package user

import (
	"context"
	"errors"
	"testing"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/usecase/mocks"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testManager struct {
	doCalled bool
}

func (m *testManager) Do(ctx context.Context, f func(ctx context.Context) error) error {
	m.doCalled = true
	return f(ctx)
}

func (m *testManager) DoWithSettings(ctx context.Context, _ trm.Settings, f func(ctx context.Context) error) error {
	m.doCalled = true
	return f(ctx)
}

func newUseCaseWithMocks(t *testing.T) (*UseCase, *mocks.MockUserRepository, *mocks.MockPullRequestRepository, *mocks.MockTeamRepository, *testManager) {
	t.Helper()

	userRepo := mocks.NewMockUserRepository(t)
	prRepo := mocks.NewMockPullRequestRepository(t)
	teamRepo := mocks.NewMockTeamRepository(t)
	trManager := &testManager{}

	uc := NewUseCase(userRepo, prRepo, teamRepo, trManager)
	return uc, userRepo, prRepo, teamRepo, trManager
}

func TestUseCase_SetIsActive_Success(t *testing.T) {
	uc, userRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	expectedUser := &entity.User{
		ID:       "u1",
		Username: "user1",
		IsActive: true,
	}

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(true, nil)

	userRepo.EXPECT().
		SetIsActive(ctx, userID, true).
		Return(nil)

	userRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(expectedUser, nil)

	result, err := uc.SetIsActive(ctx, userID, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser, result)
}

func TestUseCase_SetIsActive_UserNotFound(t *testing.T) {
	uc, userRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(false, nil)

	result, err := uc.SetIsActive(ctx, userID, true)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNotFound))
	userRepo.AssertNotCalled(t, "SetIsActive", mock.Anything, mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUseCase_SetIsActive_CheckUserExistsError(t *testing.T) {
	uc, userRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	expectedErr := errors.New("check exists failed")

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(false, expectedErr)

	result, err := uc.SetIsActive(ctx, userID, true)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	userRepo.AssertNotCalled(t, "SetIsActive", mock.Anything, mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUseCase_GetReviewList_Success(t *testing.T) {
	uc, userRepo, prRepo, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(true, nil)

	pullRequests := []*entity.PullRequest{
		{ID: "pr-1", Name: "PR 1", AuthorID: "a1", Status: entity.StatusOpen},
		{ID: "pr-2", Name: "PR 2", AuthorID: "a2", Status: entity.StatusMerged},
	}

	prRepo.EXPECT().
		GetPullRequestsByReviewerID(ctx, userID).
		Return(pullRequests, nil)

	result, err := uc.GetReviewList(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, pullRequests, result.PullRequests)
}

func TestUseCase_GetReviewList_UserNotFound(t *testing.T) {
	uc, userRepo, prRepo, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(false, nil)

	result, err := uc.GetReviewList(ctx, userID)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNotFound))
	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerID", mock.Anything, mock.Anything)
}

func TestUseCase_GetReviewList_CheckUserExistsError(t *testing.T) {
	uc, userRepo, prRepo, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	userID := "u1"

	expectedErr := errors.New("check exists failed")

	userRepo.EXPECT().
		CheckUserExists(ctx, userID).
		Return(false, expectedErr)

	result, err := uc.GetReviewList(ctx, userID)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerID", mock.Anything, mock.Anything)
}

func TestUseCase_GetStatistics_Success(t *testing.T) {
	uc, userRepo, prRepo, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()

	userIDs := []string{"u1", "u2"}

	userRepo.EXPECT().
		GetAllUsersIDs(ctx).
		Return(userIDs, nil)

	openMap := map[string]int{
		"u1": 1,
	}

	mergedMap := map[string]int{
		"u2": 2,
	}

	authorMap := map[string]int{
		"u1": 3,
		"u2": 4,
	}

	prRepo.EXPECT().
		GetOpenAndMergedReviewStatisticsForUsers(ctx).
		Return(openMap, mergedMap, nil)

	prRepo.EXPECT().
		GetAuthorStatisticsForUsers(ctx).
		Return(authorMap, nil)

	result, err := uc.GetStatistics(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.UsersStat, 2)

	u1Stat, ok := result.UsersStat["u1"]
	assert.True(t, ok)
	assert.Equal(t, "u1", u1Stat.UserID)
	assert.Equal(t, 3, u1Stat.AuthoredPRCount)
	assert.Equal(t, 1, u1Stat.OnReviewPRCount)
	assert.Equal(t, 0, u1Stat.ReviewedPRCount)

	u2Stat, ok := result.UsersStat["u2"]
	assert.True(t, ok)
	assert.Equal(t, "u2", u2Stat.UserID)
	assert.Equal(t, 4, u2Stat.AuthoredPRCount)
	assert.Equal(t, 0, u2Stat.OnReviewPRCount)
	assert.Equal(t, 2, u2Stat.ReviewedPRCount)
}

func TestUseCase_GetStatistics_GetAllUsersError(t *testing.T) {
	uc, userRepo, prRepo, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()

	expectedErr := errors.New("get users failed")

	userRepo.EXPECT().
		GetAllUsersIDs(ctx).
		Return(nil, expectedErr)

	result, err := uc.GetStatistics(ctx)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	prRepo.AssertNotCalled(t, "GetOpenAndMergedReviewStatisticsForUsers", mock.Anything)
	prRepo.AssertNotCalled(t, "GetAuthorStatisticsForUsers", mock.Anything)
}

func TestUseCase_MassDeactivateUsers_DuplicateIDs(t *testing.T) {
	uc, userRepo, prRepo, teamRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MassDeactivateUsersRequest{
		TeamName: "team-1",
		UserIDs:  []string{"u1", "u1"},
	}

	result, err := uc.MassDeactivateUsers(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrDuplicateUserIDs))
	assert.False(t, trManager.doCalled)

	teamRepo.AssertNotCalled(t, "CheckTeamNameExists", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetUsersByIDs", mock.Anything, mock.Anything)
	prRepo.AssertNotCalled(t, "GetReviewsByReviewerIDs", mock.Anything, mock.Anything)
}

func TestUseCase_MassDeactivateUsers_Success_NoReviews(t *testing.T) {
	uc, userRepo, prRepo, teamRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MassDeactivateUsersRequest{
		TeamName: "team-1",
		UserIDs:  []string{"u1"},
	}

	teamRepo.EXPECT().
		CheckTeamNameExists(ctx, req.TeamName).
		Return(true, nil)

	userRepo.EXPECT().
		GetUsersByIDs(ctx, req.UserIDs).
		Return([]*entity.User{
			{ID: "u1", TeamName: req.TeamName},
		}, nil)

	userRepo.EXPECT().
		GetActiveUsersIDsByTeamName(ctx, req.TeamName, req.UserIDs).
		Return([]string{"u2"}, nil)

	prRepo.EXPECT().
		GetReviewsByReviewerIDs(ctx, req.UserIDs).
		Return([]*entity.ReviewRecord{}, nil)

	userRepo.EXPECT().
		DeactivateUsers(ctx, req.UserIDs).
		Return(nil)

	result, err := uc.MassDeactivateUsers(ctx, req)

	assert.NoError(t, err)
	assert.True(t, trManager.doCalled)
	assert.NotNil(t, result)
	assert.Equal(t, req.TeamName, result.TeamName)
	assert.Empty(t, result.ReviewReassignments)
	assert.Empty(t, result.UnreassignedReviews)

	prRepo.AssertNotCalled(t, "RemoveReviewersBatch", mock.Anything, mock.Anything)
	prRepo.AssertNotCalled(t, "AddReviewersBatch", mock.Anything, mock.Anything)
}

func TestUseCase_MassDeactivateUsers_TeamNotFound(t *testing.T) {
	uc, userRepo, prRepo, teamRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MassDeactivateUsersRequest{
		TeamName: "unknown-team",
		UserIDs:  []string{"u1"},
	}

	teamRepo.EXPECT().
		CheckTeamNameExists(ctx, req.TeamName).
		Return(false, nil)

	result, err := uc.MassDeactivateUsers(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNotFound))
	assert.True(t, trManager.doCalled)

	userRepo.AssertNotCalled(t, "GetUsersByIDs", mock.Anything, mock.Anything)
	prRepo.AssertNotCalled(t, "GetReviewsByReviewerIDs", mock.Anything, mock.Anything)
}
