package usecase

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.uber.org/zap"
)

type TeamUseCase struct {
	teamRepo   TeamRepository
	userRepo   UserRepository
	transactor trm.Manager
	logger     *zap.Logger
}

func NewTeamUseCase(teamRepo TeamRepository, userRepo UserRepository, transactor trm.Manager, logger *zap.Logger) *TeamUseCase {
	return &TeamUseCase{
		teamRepo:   teamRepo,
		userRepo:   userRepo,
		transactor: transactor,
		logger:     logger,
	}
}

func (uc *TeamUseCase) CreateTeam(ctx context.Context, team *entity.Team) error {
	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		exists, err := uc.teamRepo.CheckTeamNameExists(ctx, team.Name)
		if err != nil {
			uc.logger.Warn("failed to check team existence", zap.Error(err))
			return err
		}
		if exists {
			return entity.ErrAlreadyExists
		}

		if err := uc.teamRepo.Create(ctx, team); err != nil {
			return err
		}

		ids := make([]string, 0, len(team.Members))
		for _, member := range team.Members {
			if member != nil {
				ids = append(ids, member.ID)
			}
		}

		existingIDs, err := uc.userRepo.FindExistingByIDs(ctx, ids)
		if err != nil {
			uc.logger.Warn("failed to find existing user IDs", zap.Error(err))
			return err
		}

		existingMembers := make([]*entity.User, 0)
		newMembers := make([]*entity.User, 0)
		for _, member := range team.Members {
			if member != nil {
				member.TeamName = team.Name
				if _, exists := existingIDs[member.ID]; exists {
					existingMembers = append(existingMembers, member)
				} else {
					newMembers = append(newMembers, member)
				}
			}
		}

		if len(existingMembers) > 0 {
			if err := uc.userRepo.UpdateUsers(ctx, existingMembers); err != nil {
				uc.logger.Warn("failed to update existing users", zap.Error(err))
				return err
			}
		}

		if len(newMembers) > 0 {
			if err := uc.userRepo.CreateBatch(ctx, newMembers); err != nil {
				uc.logger.Warn("failed to create new users", zap.Error(err))
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (uc *TeamUseCase) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	team, err := uc.teamRepo.GetByName(ctx, name)
	if err != nil {
		uc.logger.Warn("failed to get team", zap.Error(err))
		return nil, err
	}

	members, err := uc.userRepo.GetByTeamName(ctx, team.Name)
	if err != nil {
		uc.logger.Warn("failed to get team members", zap.Error(err))
		return nil, err
	}

	team.Members = members
	return team, nil
}
