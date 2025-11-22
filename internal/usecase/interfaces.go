package usecase

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type TransactionManager interface {
	manager.Manager
}

type TeamRepository interface {
	CheckTeamNameExists(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type UserRepository interface {
	CheckUserExists(ctx context.Context, id string) (bool, error)
	CreateBatch(ctx context.Context, members []*entity.Member, teamName string) error
	GetByTeamName(ctx context.Context, teamName string) ([]*entity.Member, error)
	FindExistingByIDs(ctx context.Context, ids []string) (map[string]struct{}, error)
	UpdateMembers(ctx context.Context, members []*entity.Member, teamName string) error
	SetIsActive(ctx context.Context, id string, isActive bool) error
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetReviewersForPullRequest(ctx context.Context, teamName, excludeUserID string) ([]*entity.User, error)
	GetReplacementReviewerForPullRequest(ctx context.Context, teamName string, excludeUserIDs []string) (*entity.User, error)
	GetAllUsersIDs(ctx context.Context) ([]string, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
	GetActiveUsersIDsByTeamName(ctx context.Context, teamName string, excludeIDs []string) ([]string, error)
	DeactivateUsers(ctx context.Context, ids []string) error
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	CheckPullRequestIDExists(ctx context.Context, id string) (bool, error)
	AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error)
	GetReviewersByPullRequestID(ctx context.Context, id string) ([]string, error)
	GetPullRequestsByReviewerID(ctx context.Context, userID string) ([]*entity.PullRequest, error)
	MergePullRequestByID(ctx context.Context, id string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	AddNewReviewer(ctx context.Context, prID, userID string) error
	GetOpenAndMergedReviewStatisticsForUsers(ctx context.Context) (map[string]int, map[string]int, error) // first map is open PRs, second is merged PRs
	GetAuthorStatisticsForUsers(ctx context.Context) (map[string]int, error)
	GetReviewsByReviewerIDs(ctx context.Context, reviewerIDs []string) ([]*entity.ReviewRecord, error)
	GetReviewersByPullRequestIDs(ctx context.Context, prIDs []string) (map[string][]string, error)
	RemoveReviewersBatch(ctx context.Context, records []*entity.ReviewRecord) error
	AddReviewersBatch(ctx context.Context, records []*entity.ReviewRecord) error
}
