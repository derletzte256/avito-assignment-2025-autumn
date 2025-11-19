package usecase

import (
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"fmt"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type TeamUseCase struct {
	teamRepo   TeamRepository
	userRepo   UserRepository
	transactor trm.Manager
}

func NewTeamUseCase(teamRepo TeamRepository, userRepo UserRepository, transactor trm.Manager) *TeamUseCase {
	return &TeamUseCase{
		teamRepo:   teamRepo,
		userRepo:   userRepo,
		transactor: transactor,
	}
}

func (uc *TeamUseCase) CreateTeam(ctx context.Context, team *entity.Team) error {
	err := uc.transactor.Do(ctx, func(ctx context.Context) error {
		if err := uc.teamRepo.Create(ctx, team); err != nil {
			return err
		}

		for _, member := range team.Members {
			if member == nil {
				continue
			}
			member.TeamID = team.ID
			member.TeamName = team.Name
		}

		if err := uc.userRepo.CreateBatch(ctx, team.Members); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("create team: %w", err)
	}
	return nil
}

func (uc *TeamUseCase) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	team, err := uc.teamRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get by name: %w", err)
	}

	members, err := uc.userRepo.GetByTeamID(ctx, team.ID)
	if err != nil {
		return nil, fmt.Errorf("get users by team id: %w", err)
	}

	team.Members = members
	return team, nil
}
