package team

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

func newUseCaseWithMocks(t *testing.T) (*UseCase, *mocks.MockTeamRepository, *mocks.MockUserRepository, *testManager) {
	t.Helper()

	teamRepo := mocks.NewMockTeamRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	trManager := &testManager{}

	uc := NewUseCase(teamRepo, userRepo, trManager)
	return uc, teamRepo, userRepo, trManager
}

func TestUseCase_CreateTeam_DuplicateMemberIDs(t *testing.T) {
	uc, teamRepo, userRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	member1 := &entity.Member{ID: "u1", Username: "user1", IsActive: true}
	member2 := &entity.Member{ID: "u1", Username: "user2", IsActive: true}

	team := &entity.Team{
		Name:    "team-1",
		Members: []*entity.Member{member1, member2},
	}

	err := uc.CreateTeam(ctx, team)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrDuplicateUserIDs))

	assert.False(t, trManager.doCalled)
	teamRepo.AssertNotCalled(t, "CheckTeamNameExists", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "FindExistingByIDs", mock.Anything, mock.Anything)
}

func TestUseCase_CreateTeam_TeamAlreadyExists(t *testing.T) {
	uc, teamRepo, userRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	member := &entity.Member{ID: "u1", Username: "user1", IsActive: true}
	team := &entity.Team{
		Name:    "team-1",
		Members: []*entity.Member{member},
	}

	teamRepo.EXPECT().
		CheckTeamNameExists(ctx, team.Name).
		Return(true, nil)

	err := uc.CreateTeam(ctx, team)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrAlreadyExists))
	assert.True(t, trManager.doCalled)

	teamRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "FindExistingByIDs", mock.Anything, mock.Anything)
}

func TestUseCase_CreateTeam_CheckTeamNameExistsError(t *testing.T) {
	uc, teamRepo, _, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	member := &entity.Member{ID: "u1", Username: "user1", IsActive: true}
	team := &entity.Team{
		Name:    "team-1",
		Members: []*entity.Member{member},
	}

	expectedErr := errors.New("check name failed")

	teamRepo.EXPECT().
		CheckTeamNameExists(ctx, team.Name).
		Return(false, expectedErr)

	err := uc.CreateTeam(ctx, team)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, trManager.doCalled)
}

func TestUseCase_CreateTeam_Success(t *testing.T) {
	uc, teamRepo, userRepo, trManager := newUseCaseWithMocks(t)

	ctx := context.Background()
	memberExisting := &entity.Member{ID: "u1", Username: "user1", IsActive: true}
	memberNew := &entity.Member{ID: "u2", Username: "user2", IsActive: true}

	team := &entity.Team{
		Name:    "team-1",
		Members: []*entity.Member{memberExisting, memberNew},
	}

	ids := []string{"u1", "u2"}
	existingIDs := map[string]struct{}{
		"u1": {},
	}

	teamRepo.EXPECT().
		CheckTeamNameExists(ctx, team.Name).
		Return(false, nil)

	teamRepo.EXPECT().
		Create(ctx, team).
		Return(nil)

	userRepo.EXPECT().
		FindExistingByIDs(ctx, ids).
		Return(existingIDs, nil)

	userRepo.EXPECT().
		UpdateMembers(ctx, []*entity.Member{memberExisting}, team.Name).
		Return(nil)

	userRepo.EXPECT().
		CreateBatch(ctx, []*entity.Member{memberNew}, team.Name).
		Return(nil)

	err := uc.CreateTeam(ctx, team)

	assert.NoError(t, err)
	assert.True(t, trManager.doCalled)
}

func TestUseCase_GetByName(t *testing.T) {
	type args struct {
		name            string
		teamFromRepo    *entity.Team
		teamRepoErr     error
		membersFromRepo []*entity.Member
		membersRepoErr  error
	}

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		errMatcher func(error) bool
	}{
		{
			name: "success",
			args: args{
				name: "team-1",
				teamFromRepo: &entity.Team{
					Name: "team-1",
				},
				membersFromRepo: []*entity.Member{
					{ID: "u1", Username: "user1", IsActive: true},
					{ID: "u2", Username: "user2", IsActive: false},
				},
			},
			wantErr: false,
		},
		{
			name: "team repo error",
			args: args{
				name:         "team-1",
				teamFromRepo: nil,
				teamRepoErr:  errors.New("team repo error"),
			},
			wantErr: true,
		},
		{
			name: "members repo error",
			args: args{
				name: "team-1",
				teamFromRepo: &entity.Team{
					Name: "team-1",
				},
				membersFromRepo: nil,
				membersRepoErr:  errors.New("members repo error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, teamRepo, userRepo, _ := newUseCaseWithMocks(t)
			ctx := context.Background()

			teamRepo.EXPECT().
				GetByName(ctx, tt.args.name).
				Return(tt.args.teamFromRepo, tt.args.teamRepoErr)

			if tt.args.teamRepoErr == nil {
				userRepo.EXPECT().
					GetByTeamName(ctx, tt.args.teamFromRepo.Name).
					Return(tt.args.membersFromRepo, tt.args.membersRepoErr)
			}

			gotTeam, err := uc.GetByName(ctx, tt.args.name)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, gotTeam)
			assert.Equal(t, tt.args.teamFromRepo.Name, gotTeam.Name)
			assert.Equal(t, tt.args.membersFromRepo, gotTeam.Members)
		})
	}
}
