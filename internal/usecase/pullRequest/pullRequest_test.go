package pullRequest

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

func newUseCaseWithMocks(t *testing.T) (*UseCase, *mocks.MockPullRequestRepository, *mocks.MockUserRepository, *mocks.MockTeamRepository, *testManager) {
	t.Helper()

	prRepo := mocks.NewMockPullRequestRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	teamRepo := mocks.NewMockTeamRepository(t)
	trManager := &testManager{}

	uc := NewUseCase(prRepo, userRepo, teamRepo, trManager)
	return uc, prRepo, userRepo, teamRepo, trManager
}

func TestUseCase_CreatePullRequest_Success(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.CreatePullRequestRequest{
		PullRequestID: "pr-1",
		Name:          "Test PR",
		AuthorID:      "author-1",
	}

	author := &entity.User{
		ID:       "author-1",
		Username: "author",
		IsActive: true,
		TeamName: "team-1",
	}

	reviewer1 := &entity.User{ID: "rev-1", Username: "rev1", IsActive: true, TeamName: "team-1"}
	reviewer2 := &entity.User{ID: "rev-2", Username: "rev2", IsActive: true, TeamName: "team-1"}
	reviewers := []*entity.User{reviewer1, reviewer2}
	reviewerIDs := []string{"rev-1", "rev-2"}

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(false, nil)

	userRepo.EXPECT().
		GetUserByID(ctx, req.AuthorID).
		Return(author, nil)

	userRepo.EXPECT().
		GetReviewersForPullRequest(ctx, author.TeamName, author.ID).
		Return(reviewers, nil)

	prRepo.EXPECT().
		Create(ctx, mock.MatchedBy(func(pr *entity.PullRequest) bool {
			return pr.ID == req.PullRequestID &&
				pr.Name == req.Name &&
				pr.AuthorID == req.AuthorID &&
				pr.Status == entity.StatusOpen
		})).
		Return(nil)

	prRepo.EXPECT().
		AssignReviewers(ctx, req.PullRequestID, reviewerIDs).
		Return(nil)

	createdPR := &entity.PullRequest{
		ID:       req.PullRequestID,
		Name:     req.Name,
		AuthorID: req.AuthorID,
		Status:   entity.StatusOpen,
	}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(createdPR, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewerIDs, nil)

	result, err := uc.CreatePullRequest(ctx, req)

	assert.NoError(t, err)
	assert.True(t, trManager.doCalled)
	assert.NotNil(t, result)
	assert.Equal(t, createdPR.ID, result.ID)
	assert.Equal(t, createdPR.Name, result.Name)
	assert.Equal(t, createdPR.AuthorID, result.AuthorID)
	assert.Equal(t, createdPR.Status, result.Status)
	assert.Equal(t, reviewerIDs, result.Reviewers)
}

func TestUseCase_CreatePullRequest_AlreadyExists(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.CreatePullRequestRequest{
		PullRequestID: "pr-1",
		Name:          "Test PR",
		AuthorID:      "author-1",
	}

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(true, nil)

	result, err := uc.CreatePullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrAlreadyExists))
	assert.True(t, trManager.doCalled)
	userRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUseCase_CreatePullRequest_CheckExistsError(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.CreatePullRequestRequest{
		PullRequestID: "pr-1",
		Name:          "Test PR",
		AuthorID:      "author-1",
	}

	expectedErr := errors.New("check exists failed")

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(false, expectedErr)

	result, err := uc.CreatePullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, trManager.doCalled)
	userRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUseCase_MergePullRequest_Success(t *testing.T) {
	uc, prRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MergePullRequestRequest{
		PullRequestID: "pr-1",
	}

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(true, nil)

	prRepo.EXPECT().
		MergePullRequestByID(ctx, req.PullRequestID).
		Return(nil)

	mergedPR := &entity.PullRequest{
		ID:       req.PullRequestID,
		Name:     "Merged PR",
		AuthorID: "author-1",
		Status:   entity.StatusMerged,
	}
	reviewerIDs := []string{"rev-1", "rev-2"}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(mergedPR, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewerIDs, nil)

	result, err := uc.MergePullRequest(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mergedPR.ID, result.ID)
	assert.Equal(t, mergedPR.Status, result.Status)
	assert.Equal(t, reviewerIDs, result.Reviewers)
}

func TestUseCase_MergePullRequest_NotFound(t *testing.T) {
	uc, prRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MergePullRequestRequest{
		PullRequestID: "pr-1",
	}

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(false, nil)

	result, err := uc.MergePullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNotFound))
}

func TestUseCase_MergePullRequest_CheckExistsError(t *testing.T) {
	uc, prRepo, _, _, _ := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.MergePullRequestRequest{
		PullRequestID: "pr-1",
	}

	expectedErr := errors.New("check exists failed")

	prRepo.EXPECT().
		CheckPullRequestIDExists(ctx, req.PullRequestID).
		Return(false, expectedErr)

	result, err := uc.MergePullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestUseCase_ReassignPullRequest_Success(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.ReassignPullRequestRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-1",
	}

	pullRequest := &entity.PullRequest{
		ID:       req.PullRequestID,
		AuthorID: "author-1",
		Status:   entity.StatusOpen,
	}

	oldUser := &entity.User{
		ID:       "old-1",
		Username: "old",
		IsActive: true,
		TeamName: "team-1",
	}

	reviewersBefore := []string{"old-1", "other-1"}
	newReviewer := &entity.User{
		ID:       "new-1",
		Username: "new",
		IsActive: true,
		TeamName: "team-1",
	}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(pullRequest, nil)

	userRepo.EXPECT().
		GetUserByID(ctx, req.OldUserID).
		Return(oldUser, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewersBefore, nil)

	userRepo.EXPECT().
		GetReplacementReviewerForPullRequest(ctx, oldUser.TeamName, mock.Anything).
		Return(newReviewer, nil)

	prRepo.EXPECT().
		RemoveReviewer(ctx, req.PullRequestID, oldUser.ID).
		Return(nil)

	prRepo.EXPECT().
		AddNewReviewer(ctx, req.PullRequestID, newReviewer.ID).
		Return(nil)

	updatedPR := &entity.PullRequest{
		ID:       req.PullRequestID,
		AuthorID: "author-1",
		Status:   entity.StatusOpen,
	}
	reviewersAfter := []string{"new-1", "other-1"}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(updatedPR, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewersAfter, nil)

	result, err := uc.ReassignPullRequest(ctx, req)

	assert.NoError(t, err)
	assert.True(t, trManager.doCalled)
	assert.NotNil(t, result)
	assert.NotNil(t, result.PullRequestResponse)
	assert.Equal(t, updatedPR.ID, result.PR.ID)
	assert.Equal(t, newReviewer.ID, result.ReplacedBy)
}

func TestUseCase_ReassignPullRequest_PRMerged(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.ReassignPullRequestRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-1",
	}

	pullRequest := &entity.PullRequest{
		ID:       req.PullRequestID,
		AuthorID: "author-1",
		Status:   entity.StatusMerged,
	}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(pullRequest, nil)

	result, err := uc.ReassignPullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrPRMerged))
	assert.True(t, trManager.doCalled)
	userRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUseCase_ReassignPullRequest_NotAssignedReviewer(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.ReassignPullRequestRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-1",
	}

	pullRequest := &entity.PullRequest{
		ID:       req.PullRequestID,
		AuthorID: "author-1",
		Status:   entity.StatusOpen,
	}

	oldUser := &entity.User{
		ID:       "old-1",
		Username: "old",
		IsActive: true,
		TeamName: "team-1",
	}

	reviewers := []string{"other-1"}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(pullRequest, nil)

	userRepo.EXPECT().
		GetUserByID(ctx, req.OldUserID).
		Return(oldUser, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewers, nil)

	result, err := uc.ReassignPullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNotAssignedReviewer))
	assert.True(t, trManager.doCalled)
	userRepo.AssertNotCalled(t, "GetReplacementReviewerForPullRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestUseCase_ReassignPullRequest_NoCandidate(t *testing.T) {
	uc, prRepo, userRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	req := &entity.ReassignPullRequestRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-1",
	}

	pullRequest := &entity.PullRequest{
		ID:       req.PullRequestID,
		AuthorID: "author-1",
		Status:   entity.StatusOpen,
	}

	oldUser := &entity.User{
		ID:       "old-1",
		Username: "old",
		IsActive: true,
		TeamName: "team-1",
	}

	reviewers := []string{"old-1"}

	prRepo.EXPECT().
		GetPullRequestByID(ctx, req.PullRequestID).
		Return(pullRequest, nil)

	userRepo.EXPECT().
		GetUserByID(ctx, req.OldUserID).
		Return(oldUser, nil)

	prRepo.EXPECT().
		GetReviewersByPullRequestID(ctx, req.PullRequestID).
		Return(reviewers, nil)

	userRepo.EXPECT().
		GetReplacementReviewerForPullRequest(ctx, oldUser.TeamName, mock.Anything).
		Return(nil, entity.ErrNotFound)

	result, err := uc.ReassignPullRequest(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrNoCandidate))
	assert.True(t, trManager.doCalled)
}
