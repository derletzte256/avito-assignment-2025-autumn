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
	Create(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type UserRepository interface {
	CreateBatch(ctx context.Context, users []*entity.User) error
	GetByTeamID(ctx context.Context, teamID int64) ([]*entity.User, error)
}
