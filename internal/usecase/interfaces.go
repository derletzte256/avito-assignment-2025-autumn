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
	CreateBatch(ctx context.Context, users []*entity.User) error
	GetByTeamName(ctx context.Context, teamName string) ([]*entity.User, error)
	FindExistingByIDs(ctx context.Context, ids []string) (map[string]struct{}, error)
	UpdateUsers(ctx context.Context, users []*entity.User) error
	SetIsActive(ctx context.Context, id string, isActive bool) error
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetReviewersForPullRequest(ctx context.Context, teamName, excludeUserID string) ([]*entity.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	CheckPullRequestIDExists(ctx context.Context, id string) (bool, error)
	AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error)
	GetReviewersByPullRequestID(ctx context.Context, id string) ([]string, error)
	GetPullRequestsByReviewerID(ctx context.Context, userID string) ([]*entity.PullRequest, error)
	MergePullRequestByID(ctx context.Context, id string) error
}
