package team

import (
	"avito-assignment-2025-autumn/internal/entity"
	"avito-assignment-2025-autumn/internal/usecase"
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.uber.org/zap"
)

type UseCase struct {
	teamRepo   usecase.TeamRepository
	userRepo   usecase.UserRepository
	transactor trm.Manager
	logger     *zap.Logger
}

func NewUseCase(teamRepo usecase.TeamRepository, userRepo usecase.UserRepository, transactor trm.Manager, logger *zap.Logger) *UseCase {
	return &UseCase{
		teamRepo:   teamRepo,
		userRepo:   userRepo,
		transactor: transactor,
		logger:     logger,
	}
}

func checkMemberIDsUnique(members []*entity.Member) bool {
	idsMap := make(map[string]struct{})
	for _, member := range members {
		if member != nil {
			if _, exists := idsMap[member.ID]; exists {
				return false
			}
			idsMap[member.ID] = struct{}{}
		}
	}
	return true
}

func (uc *UseCase) CreateTeam(ctx context.Context, team *entity.Team) error {
	if !checkMemberIDsUnique(team.Members) {
		return entity.ErrDuplicateUserIDs
	}

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

		existingMembers := make([]*entity.Member, 0)
		newMembers := make([]*entity.Member, 0)
		for _, member := range team.Members {
			if member != nil {
				if _, exists := existingIDs[member.ID]; exists {
					existingMembers = append(existingMembers, member)
				} else {
					newMembers = append(newMembers, member)
				}
			}
		}

		if len(existingMembers) > 0 {
			if err := uc.userRepo.UpdateMembers(ctx, existingMembers, team.Name); err != nil {
				uc.logger.Warn("failed to update existing users", zap.Error(err))
				return err
			}
		}

		if len(newMembers) > 0 {
			if err := uc.userRepo.CreateBatch(ctx, newMembers, team.Name); err != nil {
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

func (uc *UseCase) GetByName(ctx context.Context, name string) (*entity.Team, error) {
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
