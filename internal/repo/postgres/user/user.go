package user

import (
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	checkUserExistsQuery = `
		SELECT 1 FROM "user" WHERE id = $1
		`
	createUserQuery = `
		INSERT INTO "user" (id, username, is_active, team_name) 
		VALUES ($1, $2, $3, $4)
		`
	getMembersByTeamNameQuery = `
		SELECT id, username, is_active 
		FROM "user" 
		WHERE team_name = $1
		`
	findExistingByIDsQuery = `
		SELECT id 
		FROM "user" 
		WHERE id = ANY($1)
		`
	updateUsersQuery = `
		UPDATE "user" 
		SET username = $2, is_active = $3, team_name = $4 
		WHERE id = $1
		`
	setIsActiveQuery = `
		UPDATE "user" 
		SET is_active = $2 
		WHERE id = $1
		`
	getUserByIDQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user" 
		WHERE id = $1
		`
	getReviewersForPRQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user" 
		WHERE team_name = $1 AND id <> $2 AND is_active = true 
		ORDER BY RANDOM() LIMIT 2
		`
	getReviewerForPRQuery = `
		SELECT id, username, is_active, team_name 
		FROM "user"
		WHERE team_name = $1 AND id <> ALL ($2) AND is_active = true 
		ORDER BY RANDOM() LIMIT 1
		`
	getAllUsersIDsQuery = `
		SELECT id
		FROM "user"
		`
	getUsersByIDsQuery = `
		SELECT id, username, is_active, team_name
		FROM "user"
		WHERE id = ANY($1)
		`
	getActiveUsersIDsByTeamNameQuery = `
		SELECT id
		FROM "user"
		WHERE team_name = $1 AND is_active = true AND id <> ALL($2)
		`
	deactivateUsersQuery = `
		UPDATE "user"
		SET is_active = false
		WHERE id = ANY($1)
		`
)

type Repo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewUserRepo(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repo {
	return &Repo{db: db, getter: getter}
}

func (r *Repo) CheckUserExists(ctx context.Context, id string) (bool, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var exists int
	err := conn.QueryRow(ctx, checkUserExistsQuery, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repo) CreateBatch(ctx context.Context, members []*entity.Member, teamName string) error {
	l := logger.FromCtx(ctx)
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, member := range members {
		batch.Queue(createUserQuery, member.ID, member.Username, member.IsActive, teamName)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			l.Error("error closing batch results: %v\n", zap.Error(err))
		}
	}(br)

	for range members {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) GetByTeamName(ctx context.Context, teamName string) ([]*entity.Member, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getMembersByTeamNameQuery, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*entity.Member, 0)

	for rows.Next() {
		member := &entity.Member{}
		err = rows.Scan(&member.ID, &member.Username, &member.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return members, nil
}

func (r *Repo) FindExistingByIDs(ctx context.Context, ids []string) (map[string]struct{}, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, findExistingByIDsQuery, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingIDs := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		existingIDs[id] = struct{}{}
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return existingIDs, nil
}

func (r *Repo) UpdateMembers(ctx context.Context, members []*entity.Member, teamName string) error {
	l := logger.FromCtx(ctx)
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	batch := pgx.Batch{}
	for _, member := range members {
		batch.Queue(updateUsersQuery, member.ID, member.Username, member.IsActive, teamName)
	}

	br := conn.SendBatch(ctx, &batch)
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			l.Error("error closing batch results: %v\n", zap.Error(err))
		}
	}(br)

	for range members {
		result, err := br.Exec()
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return entity.ErrNotFound
		}
	}

	return nil
}

func (r *Repo) SetIsActive(ctx context.Context, id string, isActive bool) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	result, err := conn.Exec(ctx, setIsActiveQuery, id, isActive)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *Repo) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	user := &entity.User{}
	err := conn.QueryRow(ctx, getUserByIDQuery, id).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *Repo) GetReviewersForPullRequest(ctx context.Context, teamName, excludeUserID string) ([]*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getReviewersForPRQuery, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)

	for rows.Next() {
		user := &entity.User{}
		err = rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (r *Repo) GetReplacementReviewerForPullRequest(ctx context.Context, teamName string, excludeUserIDs []string) (*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	row := conn.QueryRow(ctx, getReviewerForPRQuery, teamName, excludeUserIDs)
	user := &entity.User{}
	err := row.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *Repo) GetAllUsersIDs(ctx context.Context) ([]string, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getAllUsersIDsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return ids, nil
}

func (r *Repo) GetUsersByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getUsersByIDsQuery, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)

	for rows.Next() {
		user := &entity.User{}
		if err = rows.Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (r *Repo) GetActiveUsersIDsByTeamName(ctx context.Context, teamName string, excludeIDs []string) ([]string, error) {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	rows, err := conn.Query(ctx, getActiveUsersIDsByTeamNameQuery, teamName, excludeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)

	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return ids, nil
}

func (r *Repo) DeactivateUsers(ctx context.Context, ids []string) error {
	conn := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := conn.Exec(ctx, deactivateUsersQuery, ids)
	if err != nil {
		return err
	}

	return nil
}
